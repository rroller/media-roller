# Media Roller
A mobile friendly tool for downloading videos from social media.
The backend is a Golang server that will take a URL (YouTube, Reddit, Twitter, etc),
download the video file, and return a URL to directly download the video. The video will be transcoded to produce a single mp4 file.

This is built on [yt-dlp](https://github.com/yt-dlp/yt-dlp). yt-dlp will auto update every 12 hours to make sure it's running the latest nightly build.

Note: This was written to run on a home network and should not be exposed to public traffic. There's no auth.

![Screenshot 1](https://i.imgur.com/lxwf1qU.png)

![Screenshot 2](https://i.imgur.com/TWAtM7k.png)


# Running
Make sure you have [yt-dlp](https://github.com/yt-dlp/yt-dlp) and [FFmpeg](https://github.com/FFmpeg/FFmpeg) installed then pull the repo and run:
```bash
./run.sh
```
Or for docker locally:
```bash
 ./docker-build.sh
 ./docker-run.sh
```

With Docker, published to both dockerhub and github.
* ghcr: `docker pull ghcr.io/rroller/media-roller:master`
* dockerhub: `docker pull ronnieroller/media-roller`

See:
* https://github.com/rroller/media-roller/pkgs/container/media-roller
* https://hub.docker.com/repository/docker/ronnieroller/media-roller

The files are saved to the /download directory which you can mount as needed.

## Environment variables
* `MR_DOWNLOAD_DIR` where videos are saved. Defaults to `/download`
* `MR_PROXY` will pass the value to yt-dlp witht he `--proxy` argument. Defaults to empty
* `MR_MEDIA_LIST_ENABLED` controls the Downloads library on the home page. Defaults to `true`; set to `false` to hide it.
* `MR_COOKIES_DIR` is the directory containing `cookies.txt`. Defaults to `cookies` (`/app/cookies` in Docker).

## Cookies

Place a Netscape-format `cookies.txt` export in a directory on the server. When
present, the app passes `--cookies <directory>/cookies.txt` to yt-dlp for downloads
from both the web form and API. If the file is absent, downloads run without cookies.
To refresh or remove cookies, replace or delete the file in that directory.

For local runs:

```bash
MR_COOKIES_DIR="/absolute/path/to/cookies" ./run.sh
```

For Docker, mount the directory containing `cookies.txt`:

```bash
docker run -p 3000:3000 \
  -v "$(pwd)/download:/download" \
  -v "/absolute/path/to/cookies:/app/cookies" \
  ronnieroller/media-roller
```

`./docker-run.sh` mounts `./cookies` by default; set `MR_COOKIES_DIR` to use another
host directory. The cookie file and directory must be writable by yt-dlp, which
updates its cookie jar. Keep this directory outside the download and static
folders. The default cookie directory and `*cookies*.txt` exports are excluded
from Git and Docker build contexts.

Existing explicit yt-dlp options such as `--cookies`, `--no-cookies`, and
`--cookies-from-browser` take precedence over the configured file.

# Daily dependency updates

The **Daily dependency updates** GitHub Actions workflow runs daily at 10:23 UTC
and can also be started manually from the Actions tab. It opens or refreshes a
single PR with stable Go and yt-dlp releases, Go module upgrades, and the newest
Alpine image variants available for the pinned Go and Python versions on both
amd64 and arm64. Python and FFmpeg version upgrades remain manual.

Before opening the PR, it verifies the yt-dlp checksum, runs the Go tests and
application build, and builds both container platforms without publishing.
Updates are merged through normal PR review. The workflow requires the repository
setting **Actions → General → Workflow permissions → Allow GitHub Actions to
create and approve pull requests**. Its schedule becomes active after the workflow
is merged into the default branch.

# API
To download a video directly, use the API endpoint:

```
/api/download?url=SOME_URL
```

For iOS Photos-compatible output, add `preset=ios`:

```
/api/download?preset=ios&url=SOME_URL
```

The iOS preset returns an MP4 encoded with H.264 video and AAC audio so the
Shortcuts app can save downloaded videos directly to Photos.

Create a bookmarklet, allowing one click downloads (From a PC):

```
javascript:(location.href="http://127.0.0.1:3000/fetch?url="+encodeURIComponent(location.href));
```

# Integrating with mobile
After you have your server up, install this shortcut. Update the endpoint to your server address by editing the shortcut before running it. 

https://www.icloud.com/shortcuts/12d5ab16ee3b4aa48cccfcb305ca03e4

# Unraid
media-roller is available in Unraid and can be found on the "Apps" tab by searching its name.
