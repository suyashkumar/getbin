<p align="center">
  <img src="https://suyashkumar.com/assets/img/terminal_large.png" width="80">
  <h3 align="center">GitHub Release Downloader</h3>
  <p align="center">Download latest GitHub release binaries (for your OS/arch) using <code>wget</code> or <code>curl</code></p>
  <p align="center"><code>wget --content-disposition https://bin.suyash.io/suyashkumar/ssl-proxy</code></p>
</p>

## Overview
This is a simple server (deployed @ https://bin.suyash.io) that makes it easy to download the latest binary associated with any GitHub repository release using regular old `wget` and `curl`. It attempts to use your User-Agent to fetch the right GitHub release asset for your OS/arch, but also lets you provide query parameters to specify OS/arch, and optionally uncompress the release artifact on the fly.

I mostly just built this as a way to distribute my software binaries easily without dealing with `brew`, `npm`, etc (though they certainly have their advantages & trust). I can just give my users a one line download link that will always get them the latest released binary for their platform, and _all I have to do is just update GitHub releases like I normally do_.

Basic functionality currently exists (with some assumptions, see below), but this is still a work in progress with many improvements forthcoming. This currently will work as expected with all of my repos/releases.

## Usage
Let's say you wanted to get the __latest__ [`suyashkumar/ssl-proxy`](https://github.com/suyashkumar/ssl-proxy) binary for your OS/arch. You simply:
```sh
wget -qO- "https://bin.suyash.io/suyashkumar/ssl-proxy" | tar xvz 
```
or with `curl` you usually must specify your os (since it is not included in the User-Agent):
```sh
curl -LJ "https://bin.suyash.io/suyashkumar/ssl-proxy?os=darwin" | tar xvz 
```
The request format is `GET https://bin.suyash.io/github_username/repo`

Generally, the server software attempts to detect your OS (and in the future, architecture) automatically from your `User-Agent`, but also allows you to specify your own intentions as seen with `curl` above. We're piping into `tar` here because the original release assets are compressed, but you can:

#### Uncompress on the fly
If you want to not bother with piping into tar or zip as above, the server can decompress on the fly to serve you the binary (assuming it is the only file in the archive):
```sh
wget --content-disposition "https://bin.suyash.io/suyashkumar/ssl-proxy?uncompress=true" 
```

If your release asset is __not compressed__, you can simply:
```sh
wget --content-disposition "https://bin.suyash.io/suyashkumar/ssl-proxy"
```
or
```sh
curl -LOJ "https://bin.suyash.io/suyashkumar/ssl-proxy?os=darwin"
```


## Notes & Assumptions
- If you are not using the inline uncompress feature, you'll notice that the server just transparently issues `wget` or `curl` a 301 HTTP redirect to the proper GitHub artifact URL. This way you can have some faith the artifact is whatever was uploaded to that GitHub release.
- Currently, the OS is auto matched based on the supplied User-Agent. `curl` does not supply a user agent, so the query parameter `os` must be supplied. The architecture is assumed to be `amd64`/`x86`. 
- The GitHub release asset filename currently must contain (`darwin`, `linux`, or `windows`) and (`x86` or `amd64`). More to be expanded with this basic regex in the future. Probably will want to have all `GOOS` and `GOARCH`.



## TODO
- [x] Inline, automatic uncompression of binaries
- [ ] Handle GitHub preleases
- [ ] Handle different architectures without assumptions (currently assuming x86/amd64)
- [ ] Improved binary name matching regex
- [ ] Fetch specific tagged releases

## Attribution
The [terminal icon](https://www.iconfinder.com/icons/285695/terminal_icon) used above is made by [Paomedia](https://www.iconfinder.com/paomedia) from iconfinder.com and is released under [CC BY 3.0](https://creativecommons.org/licenses/by/3.0/). 
