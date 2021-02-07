package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"manga-magnet-crawler/models"
	"manga-magnet-crawler/modules"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/nleeper/goment"
)

type Chapter struct {
	Name string
	Link string
	Date string
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	// Mongo Conntection Initialized
	dbClient := modules.MongoDBConnect(ctx)

	// Implement interface
	var mangaListModel models.MangaRepository
	mangaListModel = models.MangaListModel{dbClient.Database(os.Getenv("MONGO_DB_NAME"))}

	// Get Manga List
	for _, m := range mangaListModel.GetMangaList() {
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
		err = CreateFolder("./assets/" + m.UriName)
		if err != nil {
			log.Fatal(err)
		}

		arrChapter := []Chapter{}
		doc.Find("div#chapters > div.row").EachWithBreak(func(i int, gq *goquery.Selection) bool {
			//Skip headers row
			if i != 0 {

				//Limit
				if i <= 1 {
					rawChapterName := gq.Find("div.chapter-row > div:nth-child(2)").Text()
					chapterName := regexp.MustCompile(`(?m)\s`).ReplaceAllString(rawChapterName, "")

					link, _ := gq.Find("div.chapter-row > div:nth-child(2) > a").Attr("href")

					d, _ := gq.Find("div.chapter-row > div:nth-child(4)").Attr("title")
					dateAdded, _ := goment.New(d, "YYYY-MM-DD HH:mm:ss z (Z)")

					arrChapter = append(arrChapter, Chapter{Name: chapterName, Link: link, Date: dateAdded.Format()})

					url := os.Getenv("WEB_MANGA") + link
					res, err := http.Get(url)
					if err != nil {
						log.Fatal(err)
					}

					docPage, err := goquery.NewDocumentFromReader(res.Body)
					if err != nil {
						log.Fatal(err)
					}

					if _, err := os.Stat("./assets/" + m.UriName + "/" + chapterName); os.IsNotExist(err) {
						os.MkdirAll("./assets/"+m.UriName+"/"+chapterName, os.ModePerm)
					}

					docPage.Find("div.chapter_images-container > img").EachWithBreak(func(j int, gqDocPage *goquery.Selection) bool {
						imgLink, _ := gqDocPage.Attr("src")
						res, _ := http.Get(os.Getenv("WEB_MANGA") + imgLink)
						defer res.Body.Close()

						path := "./assets/" + m.UriName + "/" + chapterName + "/" + strconv.Itoa(j) + ".jpg"
						fmt.Println(path)
						img, _ := os.Create(path)
						defer img.Close()

						b, _ := io.Copy(img, res.Body)
						fmt.Println("File size: ", b)

						return true
					})

					return true
				}

				return false
			}
			return true
		})

		fmt.Println(arrChapter)
	}
}

func CreateFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err != nil {
			return err
		}

		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}

	}
	return nil
}
