package sync

type NotionDocument struct {
    Parent ParentInfo `json:"parent"`
    Properties NotionProperties `json:"properties"`
    Children []BulletedListItem `json:"children"`
}

type ParentInfo struct {
    DatabaseID string `json:"database_id"`
}

type NotionProperties struct {
    Name NotionName `json:"Name"`
}

type NotionName struct {
    Title []NotionTitle `json:"title"`
}

type NotionTitle struct {
    Text map[string]string `json:"text"`
    Type *string `json:"type,omitempty"`
}

type BulletedListItem struct {
    Object string `json:"object"`
    Type string `json:"type"`
    BulletedList ListItem `json:"bulleted_list_item"`
}

type ListItem struct {
    Text []NotionTitle `json:"text"`
}

func newNotionDocument(entries []string, config *Config) NotionDocument {
    children := make([]BulletedListItem, 0)

    txt := "text"
    for _, e := range entries {
        item := BulletedListItem{
                Object: "block",
                Type: "bulleted_list_item",
                BulletedList: ListItem{
                    Text: []NotionTitle{
                        {
                            Type: &txt,
                            Text: map[string]string{"content": e},
                        },
                    },
                },
            }
        children = append(children, item)
    }

    n := NotionDocument{
        Parent: ParentInfo{DatabaseID: config.DBID},
        Properties: NotionProperties{
            Name: NotionName{
                Title: []NotionTitle{
                    {
                        Text: map[string]string{
                            "content": config.DateForEntries,
                        },
                    },
                },
            },
        },
        Children: children,
    }
    return n
}
