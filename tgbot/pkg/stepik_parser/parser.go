package parser

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

type Course struct {
	Url, Title, Lessons, Duration, Level, Tests, Students,Description string
	Knowledges []string
}



func GetCourseInfo(link string, wg *sync.WaitGroup, courseInfo *Course, errChan chan<- error) {
	defer wg.Done()
	c := colly.NewCollector(
		colly.Async(true),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 5,
	})
	c.UserAgent = "Mozilla/5.0 (X11; Linux x86_64) Chrome/131.0.0.0 Safari/537.36"

	c.OnHTML("title", func(e *colly.HTMLElement) {
		courseInfo.Title = e.Text
	})

	c.OnHTML("div.course-promo__head-widget[data-type='difficulty']", func(e *colly.HTMLElement) {
		courseInfo.Level = e.Text
	})

	c.OnHTML(".course-promo-summary__students", func(e *colly.HTMLElement) {
		courseInfo.Students = e.Text
	})

	c.OnHTML(".course-promo__course-includes-aside", func(e *colly.HTMLElement) {
		e.ForEach("li", func(_ int, li *colly.HTMLElement) {
			txt := strings.TrimSpace(li.Text)
			if strings.Contains(txt, "урок") {
				courseInfo.Lessons = txt
			} else if strings.Contains(txt, "виде") {
				courseInfo.Duration = txt
			} else if strings.Contains(txt, "тест") {
				courseInfo.Tests = txt
			}
		})
	})

	if err := c.Visit(link); err != nil {
		errChan <- fmt.Errorf("error while connecting to %s: %w", link, err)
	}
	c.Wait()
}

func GetLessons(link string, wg *sync.WaitGroup, cour *Course, errChan chan<- error) {
	defer wg.Done()
	c:= colly.NewCollector(
		colly.Async(true),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 5,
	})
	c.UserAgent = "Mozilla/5.0 (X11; Linux x86_64) Chrome/131.0.0.0 Safari/537.36"
	description := " "
		c.OnHTML("section.course-promo__content-block ul.list-style__check-marks li", func(e *colly.HTMLElement) {
		cour.Knowledges = append(cour.Knowledges, e.Text)
	})
	
	c.OnHTML(".html-content", func(e *colly.HTMLElement) {
		e.ForEach("p", func(_ int, p *colly.HTMLElement) {
			description += p.Text
		})
	})
	if err := c.Visit(link); err != nil {
		errChan <- fmt.Errorf("error while connecting to %s: %w", link, err)
	}
	c.Wait()
	cour.Description = description
	
}

func Parser(link string,errCh chan<- error,courseCh chan<- Course){
	var (
		wg         sync.WaitGroup
		errChan    = make(chan error, 2)
	)
	courseInfo := Course{Url: link, Knowledges: make([]string,0,10)}
	wg.Add(2)
	go GetLessons(link, &wg, &courseInfo, errChan)
	go GetCourseInfo(link, &wg, &courseInfo, errChan)

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			fmt.Println(err)
			errCh <- err
		}
	}
	if len(courseInfo.Knowledges) == 0 {
		errCh <- fmt.Errorf("no knowledges")
	}
	fmt.Println(courseInfo)
	courseCh <- courseInfo
}
