package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	entity "github.com/HappYness-Project/ChatBackendServer/internal/entity"
)

func getTokenFromIdentityService() string {

	// This function should implement the logic to retrieve a token from the identity service.

	return "your_token_here"
}

func GetChatInfoByUserGroupId(usergroupId int, token string) entity.Chat {
	externalAPIURL := "https://example.com/api" + "/user-groups/" + strconv.Itoa(usergroupId) + "/chats"
	req, err := http.NewRequest("GET", externalAPIURL, nil)
	if err != nil {
		fmt.Println("Error creating request to external API:", err)
	} else {
		req.Header.Set("Authorization", "Bearer "+token)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error making request to external API:", err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var apiResponse map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
					fmt.Println("Error decoding external API response:", err)
				} else {
					fmt.Println("External API response:", apiResponse)
					// You can use apiResponse as needed
				}
			} else {
				fmt.Printf("External API returned status: %d\n", resp.StatusCode)
			}
		}
	}
	return entity.Chat{
		Id:          "chat_id_example",
		Type:        "group",
		UserGroupId: &usergroupId,
		ContainerId: nil,
	}
}
