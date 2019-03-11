<p align="center">
  <img src="https://suyashkumar.com/assets/img/terminal_large.png" width="80">
  <h3 align="center">GitHub Release Downloader</h3>
  <p align="center">Download latest GitHub release binaries using a single wget or curl command, (w/ optional uncompression).</p>
</p>


This software is a simple server that makes it easy to download the latest binary associated with any GitHub repository release (without having to know the GitHub release asset URL in advance). 

I mostly just built this as a way to distribute my software binaries easily without dealing with `brew`, `npm`, etc (though they certainly have their advantages & trust). I can just give my users a one line download link that will always get them the latest released binary for their platform.

## Usage
Let's say you wanted to get the latest binary for `suyashkumar/ssl-proxy` onto your machine. You could simply do:
```sh
wget -qO- "https://bin.suyash.io/suyashkumar/ssl-proxy" | tar xvz 
```
or with `curl` you usually must specify your os:
```sh
curl -LJ "https://bin.suyash.io/suyashkumar/ssl-proxy?os=darwin" | tar xvz 
```
Generally, the server software attempts to detect your OS (and in the future, architecture) automatically from your `User-Agent`, but also allows you to specify your own intentions as seen with `curl` above. 

**Note:** The above assumes that the released binary asset is compressed using `tar.gz`, but if it isn't you can leave out the piping into `tar`:

**Download GitHub release asset that is not compressed**:
```sh
curl -LOJ "https://bin.suyash.io/suyashkumar/ssl-proxy?os=darwin"
```

### Uncompress on the fly
If you want to not bother with piping into tar or zip as above, the server can decompress on the fly to serve you the binary (assuming it is the only file in the archive):
```sh
wget --content-disposition "https://bin.suyash.io/suyashkumar/ssl-proxy?os=darwin?uncompress=true" 
```

## TODO
- [x] Inline, automatic uncompression of binaries
- [ ] Handle GitHub preleases
- [ ] Handle different architectures without assumptions
- [ ] Improved binary name matching
