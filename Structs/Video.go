package Structs

import (
	"Overture-Play/Utils"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Struct

type YoutubeVideo struct {

	JSON         map[string]interface{}
	HLSFormats   []Format

	NormalFormats []Format

}

// Creator 

func CreateYoutubeVideo(Init map[string]interface{}, tokens []string) *YoutubeVideo {

	CreatedVideo := &YoutubeVideo{

		JSON:         Init,
		HLSFormats:   []Format{},
		NormalFormats: []Format{},

	}

	// Initially add any formats

	StreamingData, StreamingDataOK := Init["streamingData"].(map[string]interface{})

	if StreamingDataOK {

		Formats := []interface{}{}

		if FetchedFormats, FetchedFormatsOK := StreamingData["formats"].([]interface{}); FetchedFormatsOK {

			Formats = append(Formats, FetchedFormats...)

		}

		if FetchedAdaptiveFormats, AdaptiveFormatsOK := StreamingData["adaptiveFormats"].([]interface{}); AdaptiveFormatsOK {

			Formats = append(Formats, FetchedAdaptiveFormats...)

		}

		CreatedVideo.AddFormats(Formats, tokens)

	}

	return CreatedVideo

}

// Attached Functions

func (v *YoutubeVideo) URL() string {

	if VideoDetails, ok := v.JSON["videoDetails"].(map[string]interface{}); ok {

		if videoID, ok := VideoDetails["videoId"].(string); ok {

			return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

		}

	}

	return ""

}

func (v *YoutubeVideo) Details() VideoDetails {

	ExtractedDetails := v.JSON["videoDetails"].(map[string]interface{})

	// Safely extracts thumbnails

	var Thumbnails []Thumbnail

	if Thumb, ThumbOK := ExtractedDetails["thumbnail"].(map[string]interface{}); ThumbOK {

		if ThumbnailsData, ok := Thumb["thumbnails"].([]interface{}); ok {
			
			Thumbnails = make([]Thumbnail, 0, len(ThumbnailsData))

			for _, t := range ThumbnailsData {

				if ThumbMap, ThumbMapOK := t.(map[string]interface{}); ThumbMapOK {

					CurrentThumbnail := Thumbnail{}

					if ThumbURL, ThumbURLOK := ThumbMap["url"].(string); ThumbURLOK {

						CurrentThumbnail.URL = ThumbURL

					}

					if ThumbWidth, ThumbWidthOK := ThumbMap["width"].(float64); ThumbWidthOK {

						CurrentThumbnail.Width = int(ThumbWidth)

					}

					if ThumbHeight, ThumbHeightOK := ThumbMap["height"].(float64); ThumbHeightOK {

						CurrentThumbnail.Height = int(ThumbHeight)

					}

					Thumbnails = append(Thumbnails, CurrentThumbnail)
				}

			}

		}

	}

	// Safely extracts keywords

	var Keywords []string

	if ExtractedKeywords, KeywordsOK := ExtractedDetails["keywords"].([]interface{}); KeywordsOK {

		Keywords = make([]string, 0, len(ExtractedKeywords)) // Preallocate slice with length of ExtractedKeywords

		for _, k := range ExtractedKeywords {

			if keyword, ok := k.(string); ok {

				Keywords = append(Keywords, keyword)
				
			}
		}

	}

	// Safely extract numeric values

	var LengthSeconds int64

	if ls, ok := ExtractedDetails["lengthSeconds"].(string); ok {

		LengthSeconds, _ = strconv.ParseInt(ls, 10, 64)

	}

	var ViewCount int64

	if ExtractedViewCount, ViewCountOK := ExtractedDetails["viewCount"].(string); ViewCountOK {

		ViewCount, _ = strconv.ParseInt(ExtractedViewCount, 10, 64)

	}

	// Helpers

	GetString := func(Key string) string {

		if Str, StrOK := ExtractedDetails[Key].(string); StrOK {

			return Str

		}

		return ""

	}

	GetBool := func(Key string) bool {

		if Bln, BlnOK := ExtractedDetails[Key].(bool); BlnOK {

			return Bln

		}

		return false
	}
	
	GetFloat := func(Key string) float64 {

		if Flt, FltOK := ExtractedDetails[Key].(float64); FltOK {

			return Flt

		}

		return 0.0

	}

	CreatedDetails := VideoDetails{

		ID:                GetString("videoId"),
		URL:               v.URL(),
		Title:             GetString("title"),
		Thumbnails:        Thumbnails,
		Description:       GetString("shortDescription"),
		Duration:          LengthSeconds * 1000,
		ViewCount:         ViewCount,
		Author:            GetString("author"),
		ChannelID:         GetString("channelId"),
		Keywords:          Keywords,
		AllowRatings:      GetBool("allowRatings"),
		AverageRating:     GetFloat("averageRating"),
		IsOwnerViewing:    GetBool("isOwnerViewing"),
		IsCrawlable:       GetBool("isCrawlable"),
		IsUnpluggedCorpus: GetBool("isUnpluggedCorpus"),
		IsPrivate:         GetBool("isPrivate"),
		IsLiveContent:     GetBool("isLiveContent"),
		Formats:           v.Formats(),

	}

	return CreatedDetails

}

func (v *YoutubeVideo) Formats() []Format {

	AllFormats := append([]Format{}, v.HLSFormats...)
	AllFormats = append(AllFormats, v.NormalFormats...)

	return AllFormats

}

func (v *YoutubeVideo) AddFormats(Provided []interface{}, tokens []string) {

	for _, RawFormatInterface := range Provided {

		RawFormat, Ok := RawFormatInterface.(map[string]interface{})

		if !Ok { continue }

		ItagFloat, Ok := RawFormat["itag"].(float64)

		if !Ok { continue }

		Itag := int(ItagFloat)

		ReservedFormat, Exists := Utils.Formats[Itag]

		if !Exists { continue }

		MimeType := ReservedFormat.MimeType

		if ExtractedMimeType, ExtractedMimeOK := RawFormat["mimeType"].(string); ExtractedMimeOK {

			MimeType = ExtractedMimeType

		}

		CodecParts := strings.Split(MimeType, "\"")
		Codec := ""

		if len(CodecParts) > 1 {

			Codec = CodecParts[1]

		}

		TypeParts := strings.Split(MimeType, ";")
		FormatType := TypeParts[0]

		QualityLabel := ReservedFormat.QualityLabel

		if ExtractedQualityLabel, ExtractedQualityOK := RawFormat["qualityLabel"].(string); ExtractedQualityOK && ExtractedQualityLabel != "" {

			QualityLabel = ExtractedQualityLabel

		}

		Bitrate := ReservedFormat.Bitrate

		if ExtractedBitRate, BitRateOK := RawFormat["bitrate"].(float64); BitRateOK {

			Bitrate = int(ExtractedBitRate)

		}

		AudioBitrate := ReservedFormat.AudioBitrate

		CreatedFormat := Format{

			Itag:         Itag,
			MimeType:     MimeType,
			QualityLabel: StrToPtr(QualityLabel),
			Codec:        Codec,
			Type:         FormatType,
			Bitrate:      IntToPtr(Bitrate),
			AudioBitrate: IntToPtr(AudioBitrate),

		}

		// Other fields

		if Width, Ok := RawFormat["width"].(float64); Ok {

			W := int(Width)
			CreatedFormat.Width = &W

		}

		if Height, Ok := RawFormat["height"].(float64); Ok {

			H := int(Height)
			CreatedFormat.Height = &H

		}

		if Fps, Ok := RawFormat["fps"].(float64); Ok {

			F := int(Fps)
			CreatedFormat.Fps = &F

		}

		if Quality, Ok := RawFormat["quality"].(string); Ok {

			CreatedFormat.Quality = &Quality

		}

		if ContentLength, Ok := RawFormat["contentLength"].(string); Ok {

			Cl, _ := strconv.ParseInt(ContentLength, 10, 64)
			CreatedFormat.ContentLength = &Cl

		}

		if LastModified, Ok := RawFormat["lastModified"].(string); Ok {

			Lm, _ := strconv.ParseInt(LastModified, 10, 64)
			CreatedFormat.LastModified = &Lm

		}

		if AverageBitrate, Ok := RawFormat["averageBitrate"].(float64); Ok {

			Ab := int(AverageBitrate)
			CreatedFormat.AverageBitrate = &Ab

		}

		if ApproxDurationMs, Ok := RawFormat["approxDurationMs"].(string); Ok {

			Adm, _ := strconv.ParseInt(ApproxDurationMs, 10, 64)
			CreatedFormat.ApproxDurationMs = &Adm

		}

		if ProjectionType, Ok := RawFormat["projectionType"].(string); Ok {

			CreatedFormat.ProjectionType = &ProjectionType

		}

		// Handles InitRange and IndexRange

		if InitRange, Ok := RawFormat["initRange"].(map[string]interface{}); Ok {

			if StartStr, Ok := InitRange["start"].(string); Ok {

				if EndStr, Ok := InitRange["end"].(string); Ok {

					Start, _ := strconv.Atoi(StartStr)
					End, _ := strconv.Atoi(EndStr)
					CreatedFormat.InitRange = &Range{Start: Start, End: End}

				}

			}

		}

		if IndexRange, Ok := RawFormat["indexRange"].(map[string]interface{}); Ok {

			if StartStr, Ok := IndexRange["start"].(string); Ok {

				if EndStr, Ok := IndexRange["end"].(string); Ok {

					Start, _ := strconv.Atoi(StartStr)
					End, _ := strconv.Atoi(EndStr)
					CreatedFormat.IndexRange = &Range{Start: Start, End: End}

				}

			}

		}

		// Handles SignatureCipher

		SignatureCipher := ""

		if Sc, Ok := RawFormat["signatureCipher"].(string); Ok {
			
			SignatureCipher = Sc

		} else if Cipher, Ok := RawFormat["cipher"].(string); Ok {

			SignatureCipher = Cipher

		}

		if SignatureCipher != "" {

			CreatedFormat.SignatureCipher = &SignatureCipher

		}

		// Determine URL

		if FormatURL, Ok := RawFormat["url"].(string); Ok && SignatureCipher == "" {

			CreatedFormat.URL = FormatURL

		} else if SignatureCipher != "" {

			// Parse SignatureCipher
			Params, _ := url.ParseQuery(SignatureCipher)
			CreatedFormat.URL = Params.Get("url")

			if S := Params.Get("s"); S != "" {

				CreatedFormat.S = &S

			}

			if Sp := Params.Get("sp"); Sp != "" {

				CreatedFormat.Sp = &Sp

			}

		}

		if CreatedFormat.URL == "" { continue } // Skips if no URL is found

		// Parses URL to add some 'stuff'

		ParsedURL, Err := url.Parse(CreatedFormat.URL)

		if Err != nil { continue }

		Query := ParsedURL.Query()
		Query.Set("ratebypass", "yes")

		// Decipher signature if needed
		
		if CreatedFormat.S != nil && tokens != nil {

			Sp := "signature"

			if CreatedFormat.Sp != nil {

				Sp = *CreatedFormat.Sp

			}

			Query.Set(Sp, Utils.Decipher(tokens, *CreatedFormat.S))

		}

		ParsedURL.RawQuery = Query.Encode()
		CreatedFormat.URL = ParsedURL.String()

		v.NormalFormats = append(v.NormalFormats, *GetMetadataFromFormat(&CreatedFormat))

	}

}

// Utils 

func StrToPtr(S string) *string {
	
	if S == "" { return nil }
	return &S

}

func IntToPtr(I int) *int {
	
	if I == 0 { return nil } 
	return &I

}