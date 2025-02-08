package wp

type PostRequest struct {
    Title         string   `json:"title"`
    Content       string   `json:"content"`
    Status        string   `json:"status"`
    Slug          string   `json:"slug"`
    Categories    []int    `json:"categories"`
    FeaturedMedia int      `json:"featured_media"`
}

type PostResponse struct {
    ID      int    `json:"id"`
    Link    string `json:"link"`
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
}

type ArticleMetadata struct {
    Title     string   `json:"Title"`
    Image     string   `json:"Image"`
    Permalink string   `json:"Permalink"`
    Tag       string   `json:"Tag"`
    Category  []string `json:"Category"`
}

type Category struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type MediaResponse struct {
    ID  int    `json:"id"`
    URL string `json:"source_url"`
}