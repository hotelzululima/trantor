package main

import log "github.com/cihub/seelog"
import _ "image/png"
import _ "image/jpeg"
import _ "image/gif"

import (
	"bytes"
	"git.gitorious.org/go-pkg/epubgo.git"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"regexp"
	"strings"
)

func coverHandler(h handler) {
	vars := mux.Vars(h.r)
	if !bson.IsObjectIdHex(vars["id"]) {
		notFound(h)
		return
	}
	id := bson.ObjectIdHex(vars["id"])
	books, _, err := h.db.GetBooks(bson.M{"_id": id})
	if err != nil || len(books) == 0 {
		notFound(h)
		return
	}
	book := books[0]

	if !book.Active {
		if !h.sess.IsAdmin() {
			notFound(h)
			return
		}
	}

	fs := h.db.GetFS(FS_IMGS)
	var f *mgo.GridFile
	if vars["size"] == "small" {
		f, err = fs.OpenId(book.CoverSmall)
	} else {
		f, err = fs.OpenId(book.Cover)
	}
	if err != nil {
		log.Error("Error while opening image:", err)
		notFound(h)
		return
	}
	defer f.Close()

	headers := h.w.Header()
	headers["Content-Type"] = []string{"image/jpeg"}

	io.Copy(h.w, f)
}

func GetCover(e *epubgo.Epub, title string, db *DB) (bson.ObjectId, bson.ObjectId) {
	imgId, smallId := coverFromMetadata(e, title, db)
	if imgId != "" {
		return imgId, smallId
	}

	imgId, smallId = searchCommonCoverNames(e, title, db)
	if imgId != "" {
		return imgId, smallId
	}

	/* search for img on the text */
	exp, _ := regexp.Compile("<.*ima?g.*[(src)(href)]=[\"']([^\"']*(\\.[^\\.\"']*))[\"']")
	it, errNext := e.Spine()
	for errNext == nil {
		file, err := it.Open()
		if err != nil {
			break
		}
		defer file.Close()

		txt, err := ioutil.ReadAll(file)
		if err != nil {
			break
		}
		res := exp.FindSubmatch(txt)
		if res != nil {
			href := string(res[1])
			urlPart := strings.Split(it.URL(), "/")
			url := strings.Join(urlPart[:len(urlPart)-1], "/")
			if href[:3] == "../" {
				href = href[3:]
				url = strings.Join(urlPart[:len(urlPart)-2], "/")
			}
			href = strings.Replace(href, "%20", " ", -1)
			href = strings.Replace(href, "%27", "'", -1)
			href = strings.Replace(href, "%28", "(", -1)
			href = strings.Replace(href, "%29", ")", -1)
			if url == "" {
				url = href
			} else {
				url = url + "/" + href
			}

			img, err := e.OpenFile(url)
			if err == nil {
				defer img.Close()
				return storeImg(img, title, db)
			}
		}
		errNext = it.Next()
	}
	return "", ""
}

func coverFromMetadata(e *epubgo.Epub, title string, db *DB) (bson.ObjectId, bson.ObjectId) {
	metaList, _ := e.MetadataAttr("meta")
	for _, meta := range metaList {
		if meta["name"] == "cover" {
			img, err := e.OpenFileId(meta["content"])
			if err == nil {
				defer img.Close()
				return storeImg(img, title, db)
			}
		}
	}
	return "", ""
}

func searchCommonCoverNames(e *epubgo.Epub, title string, db *DB) (bson.ObjectId, bson.ObjectId) {
	for _, p := range []string{"cover.jpg", "Images/cover.jpg", "images/cover.jpg", "cover.jpeg", "cover1.jpg", "cover1.jpeg"} {
		img, err := e.OpenFile(p)
		if err == nil {
			defer img.Close()
			return storeImg(img, title, db)
		}
	}
	return "", ""
}

func storeImg(img io.Reader, title string, db *DB) (bson.ObjectId, bson.ObjectId) {
	/* open the files */
	fBig, err := createCoverFile(title, db)
	if err != nil {
		log.Error("Error creating", title, ":", err.Error())
		return "", ""
	}
	defer fBig.Close()

	fSmall, err := createCoverFile(title+"_small", db)
	if err != nil {
		log.Error("Error creating", title+"_small", ":", err.Error())
		return "", ""
	}
	defer fSmall.Close()

	/* resize img */
	var img2 bytes.Buffer
	img1 := io.TeeReader(img, &img2)
	jpgOptions := jpeg.Options{IMG_QUALITY}
	imgResized, err := resizeImg(img1, IMG_WIDTH_BIG)
	if err != nil {
		log.Error("Error resizing big image:", err.Error())
		return "", ""
	}
	err = jpeg.Encode(fBig, imgResized, &jpgOptions)
	if err != nil {
		log.Error("Error encoding big image:", err.Error())
		return "", ""
	}
	imgSmallResized, err := resizeImg(&img2, IMG_WIDTH_SMALL)
	if err != nil {
		log.Error("Error resizing small image:", err.Error())
		return "", ""
	}
	err = jpeg.Encode(fSmall, imgSmallResized, &jpgOptions)
	if err != nil {
		log.Error("Error encoding small image:", err.Error())
		return "", ""
	}

	idBig, _ := fBig.Id().(bson.ObjectId)
	idSmall, _ := fSmall.Id().(bson.ObjectId)
	return idBig, idSmall
}

func createCoverFile(title string, db *DB) (*mgo.GridFile, error) {
	fs := db.GetFS(FS_IMGS)
	return fs.Create(title + ".jpg")
}

func resizeImg(imgReader io.Reader, width uint) (image.Image, error) {
	img, _, err := image.Decode(imgReader)
	if err != nil {
		return nil, err
	}

	return resize.Resize(width, 0, img, resize.NearestNeighbor), nil
}
