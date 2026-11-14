#!/usr/bin/env python3
"""Refresh version pins; Go module upgrades run separately with the new toolchain."""

import hashlib
import json
import os
from pathlib import Path
import re
from urllib.parse import urlencode
from urllib.request import Request, urlopen


def download(url):
    headers = {"User-Agent": "media-roller-dependency-updater"}
    if url.startswith("https://api.github.com/") and os.environ.get("GH_TOKEN"):
        headers["Authorization"] = f"Bearer {os.environ['GH_TOKEN']}"
    with urlopen(Request(url, headers=headers), timeout=60) as response:
        return response.read()


def json_url(url):
    return json.loads(download(url))


def replace_once(text, pattern, replacement):
    result, count = re.subn(pattern, lambda _: replacement, text, flags=re.MULTILINE)
    if count != 1:
        raise ValueError(f"Expected exactly one match for {pattern!r}; got {count}")
    return result


def version_tuple(version):
    return tuple(int(part) for part in version.split("."))


def latest_alpine(image, version, minimum):
    query = urlencode({"page_size": 100, "name": f"{version}-alpine"})
    url = f"https://hub.docker.com/v2/repositories/library/{image}/tags?{query}"
    candidates = []
    while url:
        page = json_url(url)
        for tag in page["results"]:
            match = re.fullmatch(re.escape(version) + r"-alpine(\d+\.\d+)", tag["name"])
            platforms = {(item["os"], item["architecture"]) for item in tag["images"]}
            if match and {("linux", "amd64"), ("linux", "arm64")} <= platforms:
                alpine = match[1]
                if version_tuple(alpine) >= version_tuple(minimum):
                    candidates.append(alpine)
        url = page.get("next")
    if not candidates:
        raise ValueError(f"No multi-platform {image}:{version}-alpine >= {minimum}")
    return max(candidates, key=version_tuple)


def latest_yt_dlp():
    release = json_url("https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest")
    version = release["tag_name"]
    if release["prerelease"] or release["draft"] or not re.fullmatch(r"\d{4}\.\d{2}\.\d{2}", version):
        raise ValueError(f"Unexpected yt-dlp release: {version}")
    base = f"https://github.com/yt-dlp/yt-dlp/releases/download/{version}/"
    sums = download(base + "SHA2-256SUMS").decode()
    checksum = next(line.split()[0] for line in sums.splitlines()
                    if len(line.split()) == 2 and line.split()[1] in ("yt-dlp", "*yt-dlp"))
    if not re.fullmatch(r"[0-9a-f]{64}", checksum):
        raise ValueError("Invalid yt-dlp checksum")
    if hashlib.sha256(download(base + "yt-dlp")).hexdigest() != checksum:
        raise ValueError("yt-dlp checksum verification failed")
    return version, checksum


def main():
    docker_path = Path("Dockerfile")
    docker = docker_path.read_text()
    go_image = re.search(r"^FROM golang:(\d+\.\d+\.\d+)-alpine(\d+\.\d+) AS builder$", docker, re.M)
    python_image = re.search(r"^FROM python:(\d+\.\d+\.\d+)-alpine(\d+\.\d+)$", docker, re.M)
    if not go_image or not python_image:
        raise ValueError("Unrecognized Docker base image pins")
    releases = json_url("https://go.dev/dl/?mode=json")
    go_version = max((release["version"].removeprefix("go") for release in releases
                      if release["stable"] and re.fullmatch(r"go\d+\.\d+\.\d+", release["version"])),
                     key=version_tuple)
    if version_tuple(go_version) < version_tuple(go_image[1]):
        raise ValueError("Refusing to downgrade Go")
    go_alpine = latest_alpine("golang", go_version, go_image[2])
    python_alpine = latest_alpine("python", python_image[1], python_image[2])
    yt_version, checksum = latest_yt_dlp()
    current_yt = re.search(r"releases/download/(\d{4}\.\d{2}\.\d{2})/yt-dlp", docker)
    if not current_yt or yt_version < current_yt[1]:
        raise ValueError("Unrecognized yt-dlp pin or attempted downgrade")
    docker = replace_once(docker, r"^FROM golang:.* AS builder$",
                          f"FROM golang:{go_version}-alpine{go_alpine} AS builder")
    docker = replace_once(docker, r"^FROM python:.*$",
                          f"FROM python:{python_image[1]}-alpine{python_alpine}")
    docker = replace_once(docker, r"https://github\.com/yt-dlp/yt-dlp/releases/download/[^/]+/yt-dlp",
                          f"https://github.com/yt-dlp/yt-dlp/releases/download/{yt_version}/yt-dlp")
    docker = replace_once(docker, r'[0-9a-f]{64} /usr/local/bin/yt-dlp',
                          f"{checksum} /usr/local/bin/yt-dlp")
    mod_path = Path("go.mod")
    mod = replace_once(mod_path.read_text(), r"^go \S+$", f"go {go_version}")
    tools_path = Path(".tool-versions")
    tools = replace_once(tools_path.read_text(), r"^golang \S+$", f"golang {go_version}")
    # Resolve and verify every upstream artifact before changing any files.
    docker_path.write_text(docker)
    mod_path.write_text(mod)
    tools_path.write_text(tools)
    print(f"Go {go_version}; yt-dlp {yt_version}; Alpine builder {go_alpine}, runtime {python_alpine}")


if __name__ == "__main__":
    main()
