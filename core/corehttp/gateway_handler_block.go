package corehttp

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"time"

	ipath "github.com/ipfs/interface-go-ipfs-core/path"
)

// serveRawBlock returns bytes behind a raw block
func (i *gatewayHandler) serveRawBlock(ctx context.Context, w http.ResponseWriter, r *http.Request, resolvedPath ipath.Resolved, contentPath ipath.Path, begin time.Time) {
	blockCid := resolvedPath.Cid()
	blockReader, err := i.api.Block().Get(ctx, resolvedPath)
	if err != nil {
		webError(w, "ipfs block get "+blockCid.String(), err, http.StatusInternalServerError)
		return
	}
	block, err := ioutil.ReadAll(blockReader)
	if err != nil {
		webError(w, "ipfs block get "+blockCid.String(), err, http.StatusInternalServerError)
		return
	}
	content := bytes.NewReader(block)

	// Set Content-Disposition
	name := blockCid.String() + ".bin"
	setContentDispositionHeader(w, name, "attachment")

	// Set remaining headers
	modtime := addCacheControlHeaders(w, r, contentPath, blockCid)
	w.Header().Set("Content-Type", "application/vnd.ipld.raw")
	w.Header().Set("X-Content-Type-Options", "nosniff") // no funny business in the browsers :^)

	// ServeContent will take care of
	// If-None-Match+Etag, Content-Length and range requests
	ServeContent(w, r, name, modtime, content)
}
