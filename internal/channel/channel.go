package channel

type Channel struct {
	ChannelId       string   `json:"externalId"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	IsFamilySafe    bool     `json:"isFamilySafe"`
	Keywords        string   `json:"keywords"`
	Thumbnails      []string `json:"thumbnails"`
	Links           []string `json:"links"`
	ViewCount       int      `json:"viewCount"`
	TotalSubscriber int      `json:"totalSubscriber"`
	TotalVideo      int      `json:"totalVideo"`
	Joined          struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Date  int `json:"date"`
	} `json:"joined"`
}
