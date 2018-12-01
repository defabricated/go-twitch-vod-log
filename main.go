package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	videoID := os.Args[1]

	cs, err := GetLogs(videoID)
	if err != nil {
		panic(err)
		return
	}

	l := make([]string, len(cs))
	for i, c := range cs {
		t := int(c.ContentOffsetSeconds)
		h := int(math.Floor(float64(t / 3600)))
		m := int(math.Floor(float64(t % 3600) / 60))
		s := int(math.Floor(float64((t % 3600) % 60)))

		l[i] = fmt.Sprintf("[%02d:%02d:%02d] [%s] %s: %s", h, m, s, c.Source, c.Commenter.DisplayName, c.Message.Body)
	}

	if err = ioutil.WriteFile(fmt.Sprintf("%s-logs.txt", videoID), []byte(strings.Join(l, "\n")), os.ModePerm); err != nil {
		panic(err)
	}

	fmt.Printf("Successfully downloaded logs from video %s.\n", videoID)
}

func GetLogs(videoID string) (cs []*Comment, err error) {
	videoUrl := fmt.Sprintf("https://api.twitch.tv/v5/videos/%s/comments?", videoID)
	nextCursor := ""

	cs = make([]*Comment, 0)
	for len(cs) == 0 || nextCursor != "" {
		client := http.Client{}

		query := url.Values{}
		if nextCursor == "" {
			query.Add("content_offset_seconds", "0")
		} else {
			query.Add("cursor", nextCursor)
		}

		var req *http.Request
		req, err = http.NewRequest("GET", videoUrl+query.Encode(), nil)
		if err != nil {
			return
		}

		req.Header.Add("Client-ID", "jzkbprff40iqj646a697cyrvl0zt2m6")

		var resp *http.Response
		resp, err = client.Do(req)
		if err != nil {
			return
		}

		var bytes []byte
		bytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		var twitchResp *Response
		if err = json.Unmarshal(bytes, &twitchResp); err != nil {
			return
		}

		nextCursor = twitchResp.NextCursor
		cs = append(cs, twitchResp.Comments...)

		lt := int(cs[len(cs)-1].ContentOffsetSeconds)

		h := int(math.Floor(float64(lt / 3600)))
		m := int(math.Floor(float64(lt % 3600) / 60))
		s := int(math.Floor(float64((lt % 3600) % 60)))
		fmt.Printf("%02d:%02d:%02d\n", h, m, s)
	}
	return
}

type Response struct {
	Comments   []*Comment `json:"comments"`
	PrevCursor string     `json:"_prev"`
	NextCursor string     `json:"_next"`
}

type Comment struct {
	ID                   string    `json:"_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	ChannelID            string    `json:"channel_id"`
	ContentType          string    `json:"content_type"`
	ContentID            string    `json:"content_id"`
	ContentOffsetSeconds float64   `json:"content_offset_seconds"`
	Commenter            Commenter `json:"commenter"`
	Source               string    `json:"source"`
	State                string    `json:"state"`
	Message              *Message  `json:"message"`
	MoreReplies          bool      `json:"more_replies"`
}

type Commenter struct {
	ID          string    `json:"_id"`
	DisplayName string    `json:"display_name"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Bio         string    `json:"bio"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Logo        string    `json:"logo"`
}

type Message struct {
	Body      string      `json:"body"`
	Fragments []*Fragment `json:"fragments"`
	IsAction  bool        `json:"is_action"`
	UserColor string      `json:"user_color"`
}

type Fragment struct {
	Text string `json:"text"`
}
