package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/inuad/manga-magnet-crawler/models"
	"github.com/inuad/manga-magnet-crawler/modules"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	// Mongo Conntection Initialized
	dbClient := modules.MongoDBConnect(ctx)

	mangaMagnetModel := models.MangaMagnetModel{dbClient.Database(os.Getenv("MONGO_DB_NAME"))}

	// Get Manga List
	for _, m := range mangaMagnetModel.GetMangaList() {
		fmt.Println("Start download : " + m.Name)

		res, err := http.Get(m.Link)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Create manga assets folder
		err = CreateFolder(os.Getenv("STORAGE_PATH") + m.UriName + "/")
		if err != nil {
			log.Fatal(err)
		}

		doc.Find("div#chapters > div.row").EachWithBreak(func(i int, gq *goquery.Selection) bool {
			//Skip headers row
			if i != 0 {

				// Limit
				if i <= 1 {
					rawParentChapterName := gq.Find("div.chapter-row > div:nth-child(2)")
					rawChapterName := rawParentChapterName.Find("a:last-of-type > div:last-child").Text()
					chapterName := regexp.MustCompile(`(?m)\s`).ReplaceAllString(rawChapterName, "")

					r, err := mangaMagnetModel.GetLatestChapter(m.ID, chapterName)
					if err == nil {
						if chapterName == r.ChapterName {
							fmt.Println("Chapter is up to date, " + chapterName + " skip...")
							return false
						}
					}
					link, _ := gq.Find("div.chapter-row > div:nth-child(2) > a:last-of-type").Attr("href")

					// d, _ := gq.Find("div.chapter-row > div:nth-child(4)").Attr("title")
					// dateAdded, _ := goment.New(d, "YYYY-MM-DD HH:mm:ss z (Z)")

					// Request chapter page
					url := os.Getenv("WEB_MANGA") + link
					res, err := http.Get(url)
					if err != nil {
						log.Fatal(err)
					}

					docPage, err := goquery.NewDocumentFromReader(res.Body)
					if err != nil {
						log.Fatal(err)
					}

					if _, err := os.Stat(os.Getenv("STORAGE_PATH") + m.UriName + "/" + chapterName); os.IsNotExist(err) {
						os.MkdirAll(os.Getenv("STORAGE_PATH")+m.UriName+"/"+chapterName, os.ModePerm)
					}

					var arrImagePath []string
					docPage.Find("div[class*='chapter-images'] > img").EachWithBreak(func(j int, gqDocPage *goquery.Selection) bool {
						imgLink, _ := gqDocPage.Attr("src")

						res, _ := http.Get(os.Getenv("WEB_MANGA") + imgLink)
						defer res.Body.Close()

						pathWithName := chapterName + "/" + strconv.Itoa(j) + ".jpg"
						path := os.Getenv("STORAGE_PATH") + m.UriName + "/" + pathWithName
						arrImagePath = append(arrImagePath, pathWithName)
						img, _ := os.Create(path)
						defer img.Close()

						b, _ := io.Copy(img, res.Body)
						fmt.Println("File size: ", b)

						return true
					})

					chapterStruct := models.Chapter{m.ID, chapterName, link, arrImagePath, time.Now()}
					mangaMagnetModel.SetMangaChapter(&chapterStruct)
					fmt.Println(chapterName + " Done.")

					return true
				}

				return false
			}
			return true
		})

	}
}

func CreateFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
