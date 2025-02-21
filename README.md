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

## Docker Environemnt Variables
* `MR_DOWNLOAD_DIR` where videos are saved. Defaults to `/download`
* `MR_PROXY` will pass the value to yt-dlp witht he `--proxy` argument. Defaults to empty

# API
To download a video directly, use the API endpoint:

```
/api/download?url=SOME_URL
```

Create a bookmarklet, allowing one click downloads (From a PC):

```
javascript:(location.href="http://127.0.0.1:3000/fetch?url="+encodeURIComponent(location.href));
```

# Integrating with mobile
After you have your server up, install this shortcut. Update the endpoint to your server address by editing the shortcut before running it. 

https://www.icloud.com/shortcuts/d3b05b78eb434496ab28dd91e1c79615

# Unraid
media-roller is available in Unraid and can be found on the "Apps" tab by searching its name.
