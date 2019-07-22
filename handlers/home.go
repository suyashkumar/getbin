package handlers

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const homeMessage = `
<html>
	<head>
	<link rel="stylesheet" href="https://unpkg.com/purecss@1.0.0/build/base-min.css">
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/3.0.1/github-markdown.min.css">	
	</head>
	<article class="markdown-body">
	<div style="max-width:1000px;padding-left:20px;padding-right:20px;margin:auto">
		<h1> GitHub Release Downloader </h1>
		<p>
			This tool helps you download the latest binaries from GitHub releases quickly using curl or wget.
			Find out more at the <a href="https://github.com/suyashkumar/getbin">GitHub README for this project.</a>
		</p>
		<p>
			Let's say you wanted to download the latest version of	
			<a href="https://github.com/suyashkumar/ssl-proxy" target="_blank"> <code>ssl-proxy</code> </a>
			for your OS: 
		</p>
	</div>
	<div style="max-width:1000px;padding-left:20px;padding-right:20px;margin:auto;min-width:845px">
		<p>
			Download and untar the <b>latest</b> release of ssl-proxy for your OS (will be based on wget's 
			<code>User-Agent</code>): <br />
			<code>wget -qO- "https://getbin.io/suyashkumar/ssl-proxy" | tar xvz</code> <br />
		</p>
		<p>
			You can also specify the OS you wish to download for as follows (can be either <code>darwin</code>, 
			<code>linux</code>, or <code>windows</code>: <br />
			<code>wget -qO- "https://getbin.io/suyashkumar/ssl-proxy?os=darwin" | tar xvz</code> <br />
		</p> 
		<p>
			You can also let the server handle uncompression for you: <br /> 
			<code>wget --content-disposition "https://getbin.io/suyashkumar/ssl-proxy?os=darwin?uncompress=true"</code> <br />
		</p>
		<p>
			You can also use <code>curl</code>. Note, you must always specify <code>os</code> with curl. <br />
			<code> curl -LJ "https://getbin.io/suyashkumar/ssl-proxy?os=darwin" | tar xvz </code> <br />
		</p>
	</div>
	</article>
</html> 
`

// Home is the index handler that simply returns the homeMessage HTML above
func Home(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprint(w, homeMessage)
}
