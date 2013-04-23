package main

import (
	"bytes"
	"git.gitorious.org/go-pkg/epubgo.git"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"regexp"
	"strings"
)

func GetCover(e *epubgo.Epub, title string) (bson.ObjectId, bson.ObjectId) {
	imgId, smallId := searchCommonCoverNames(e, title)
	if imgId != "" {
		return imgId, smallId
	}

	/* search for img on the text */
	exp, _ := regexp.Compile("<ima?g.*[(src)(href)]=[\"']([^\"']*(\\.[^\\.\"']*))[\"']")
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
			urlPart := strings.Split(it.Url(), "/")
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
				return storeImg(img, title)
			}
		}
		errNext = it.Next()
	}
	return "", ""
}

func searchCommonCoverNames(e *epubgo.Epub, title string) (bson.ObjectId, bson.ObjectId) {
	for _, p := range []string{"cover.jpg", "Images/cover.jpg", "cover.jpeg", "cover1.jpg", "cover1.jpeg"} {
		img, err := e.OpenFile(p)
		if err == nil {
			defer img.Close()
			return storeImg(img, title)
		}
	}
	return "", ""
}

func storeImg(img io.Reader, title string) (bson.ObjectId, bson.ObjectId) {
	/* open the files */
	fBig, err := createCoverFile(title)
	if err != nil {
		log.Println("Error creating", title, ":", err.Error())
		return "", ""
	}
	defer fBig.Close()

	fSmall, err := createCoverFile(title + "_small")
	if err != nil {
		log.Println("Error creating", title+"_small", ":", err.Error())
		return "", ""
	}
	defer fSmall.Close()

	/* resize img */
	var img2 bytes.Buffer
	img1 := io.TeeReader(img, &img2)
	jpgOptions := jpeg.Options{IMG_QUALITY}
	imgResized, err := resizeImg(img1, IMG_WIDTH_BIG)
	if err != nil {
		log.Println("Error resizing big image:", err.Error())
		return "", ""
	}
	err = jpeg.Encode(fBig, imgResized, &jpgOptions)
	if err != nil {
		log.Println("Error encoding big image:", err.Error())
		return "", ""
	}
	imgSmallResized, err := resizeImg(&img2, IMG_WIDTH_SMALL)
	if err != nil {
		log.Println("Error resizing small image:", err.Error())
		return "", ""
	}
	err = jpeg.Encode(fSmall, imgSmallResized, &jpgOptions)
	if err != nil {
		log.Println("Error encoding small image:", err.Error())
		return "", ""
	}

	idBig, _ := fBig.Id().(bson.ObjectId)
	idSmall, _ := fSmall.Id().(bson.ObjectId)
	return idBig, idSmall
}

func createCoverFile(title string) (*mgo.GridFile, error) {
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