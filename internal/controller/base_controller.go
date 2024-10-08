package controller

import (
	"aniverse/internal/crawler"
	"aniverse/internal/extractor"
	"aniverse/internal/provider/anilist"
	"aniverse/internal/provider/gogoanime"
	"aniverse/internal/provider/mal"
)

type BaseController struct {
	anilist     *anilist.AniListBase
	gogoanime   *gogoanime.GogoAnime
	myanimelist *mal.MyAnimeList
	extractor   *extractor.Gogocdn
	crawler     *crawler.BaseCrawler
}

func NewBaseController() *BaseController {
	crawler := crawler.NewBaseCrawler()
	return &BaseController{
		anilist:     anilist.NewAniListBase(),
		gogoanime:   gogoanime.NewGogoAnime(),
		myanimelist: mal.NewMyAnimeList(),
		extractor:   extractor.NewGogocdn(crawler),
		crawler:     crawler,
	}
}
