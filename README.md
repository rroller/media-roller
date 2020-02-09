# Media Roller
A mobile friendly tool for downloading videos from social media.
The backend is is a Golang server that will take a URL (YouTube, Reddit, Twitter, etc),
download the video file, and return a URL to download the video. The video will be transcoced as needed to produce a single mp4 file.

Note: This is for home use only. There's no auth.

![GitHub Logo](https://i.imgur.com/lxwf1qU.png)

![GitHub Logo](https://i.imgur.com/TWAtM7k.png)


# Running
Pull the repo then run
```bash
./run.sh
```

With Docker: `ronnieroller/media-roller:latest`.
See https://hub.docker.com/repository/docker/ronnieroller/media-roller
The files will be saved to the /download directory which you can mount as needed.


With Unraid: TODO: This works with Unraid, I'm working on a template and will publish it soon.

# Integrating with mobile
I'm working on an iOS shortcut will download the video to the camera roll for a supplied URL.
