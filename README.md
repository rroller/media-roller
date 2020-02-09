# Media Roller
A mobile friendly tool for downloading videos from social media.
The backend is a Golang server that will take a URL (YouTube, Reddit, Twitter, etc),
download the video file, and return a URL to directly download the video. The video will be transcoced as needed to produce a single mp4 file.

Note: This was written to run on a home network and wasn't originally written to be exposed to public traffic. Currently there's no auth. This might change and feel free to send a pull request, but right now, keep this on your internal network and do not expose it.

![Screenshot 1](https://i.imgur.com/lxwf1qU.png)

![Screenshot 2](https://i.imgur.com/TWAtM7k.png)


# Running
Pull the repo then run
```bash
./run.sh
```

With Docker: `ronnieroller/media-roller:latest`.
See https://hub.docker.com/repository/docker/ronnieroller/media-roller
The files will be saved to the /download directory which you can mount as needed.


With Unraid: TODO: This works with Unraid, I'm working on a template and will publish it soon.

# API
To download a video directly, use the API endpoint:

```
/api/download?url=SOME_URL
```

# Integrating with mobile
I'm working on an iOS shortcut will download the video to the camera roll for a supplied URL.
