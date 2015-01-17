package main

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/NebulousLabs/Sia/sia/components"
)

func (d *daemon) fileUploadHandler(w http.ResponseWriter, req *http.Request) {
	pieces, err := strconv.Atoi(req.FormValue("pieces"))
	if err != nil {
		http.Error(w, "Malformed pieces", 400)
		return
	}
	// this is slightly dangerous, but we assume the user won't try to attack siad
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Could not read file data: "+err.Error(), 400)
		return
	}

	// TODO: is "" a valid nickname? The renter should probably prevent this.
	err = d.core.RentSmallFile(components.RentSmallFileParameters{
		FullFile:    data,
		Nickname:    req.FormValue("nickname"),
		TotalPieces: pieces,
	})
	if err != nil {
		http.Error(w, "Upload failed: "+err.Error(), 500)
		return
	}

	writeSuccess(w)
}

func (d *daemon) fileDownloadHandler(w http.ResponseWriter, req *http.Request) {
	err := d.core.RenterDownload(req.FormValue("nickname"), d.downloadDir+req.FormValue("filename"))
	if err != nil {
		// TODO: if this err is a user error (e.g. bad nickname), return 400 instead
		http.Error(w, "Download failed: "+err.Error(), 500)
		return
	}

	writeSuccess(w)
}

func (d *daemon) fileStatusHandler(w http.ResponseWriter, req *http.Request) {
	info, err := d.core.RentInfo()
	if err != nil {
		http.Error(w, "Couldn't get renter info: "+err.Error(), 500)
		return
	}

	writeJSON(w, info)
}
