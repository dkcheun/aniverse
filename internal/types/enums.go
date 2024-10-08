package types

type Format string

const (
	FormatUnknown Format = "UNKNOWN"
	FormatMovie   Format = "MOVIE"
	FormatONA     Format = "ONA"
	FormatOVA     Format = "OVA"
	FormatSpecial Format = "SPECIAL"
	FormatTV      Format = "TV"
	FormatTVShort Format = "TV_SHORT"
)

type MediaStatus string

const (
	StatusFinished       MediaStatus = "FINISHED"
	StatusReleasing      MediaStatus = "RELEASING"
	StatusNotYetReleased MediaStatus = "NOT_YET_RELEASED"
	StatusCancelled      MediaStatus = "CANCELLED"
)

type Season string

const (
	SeasonWinter  Season = "WINTER"
	SeasonSpring  Season = "SPRING"
	SeasonSummer  Season = "SUMMER"
	SeasonFall    Season = "FALL"
	SeasonUnknown Season = "UNKNOWN"
)

type MediaType string

const (
	TypeAnime MediaType = "ANIME"
	TypeManga MediaType = "MANGA"
)

type SubType string

const (
	SubTypeSub SubType = "SUB"
	SubTypeDub SubType = "DUB"
)

type StreamingServer string

const (
	ServerGogoCDN      StreamingServer = "GogoCDN"
	ServerVidStreaming StreamingServer = "VidStreaming"
	ServerStreamSB     StreamingServer = "StreamSB"
)
