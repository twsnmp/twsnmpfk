package notify

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

var lastSendLine = int64(0)
var failedToken = ""

func SendLine(c *datastore.NotifyConfEnt, message string, stickerPackageId, stickerId int) error {
	token := c.LineToken
	if token == "" || failedToken == token {
		return fmt.Errorf("invalid line token")
	}
	if message == "" {
		return fmt.Errorf("no line message")
	}

	if time.Now().Unix()-lastSendLine < 3 {
		time.Sleep(time.Second * 3)
	}

	data := url.Values{"message": {message}}
	if stickerPackageId > 0 && stickerId > 0 {
		data.Add("stickerPackageId", fmt.Sprintf("%d", stickerPackageId))
		data.Add("stickerId", fmt.Sprintf("%d", stickerId))
	}
	r, _ := http.NewRequest("POST", "https://notify-api.line.me/api/notify", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}
	if resp.StatusCode == 401 {
		log.Printf("line token is expired %s", token)
		failedToken = token
	}
	return fmt.Errorf("send line code=%d", resp.StatusCode)
}

// sendNotifyLine : Lineへ通知メッセージを送信する
func SendNotifyLine(l *datastore.EventLogEnt) {
	if datastore.NotifyConf.LineToken == "" {
		return
	}
	nl := getLevelNum(datastore.NotifyConf.LineLevel)
	if nl == 3 {
		return
	}
	if l.Level == "repair" {
		if !datastore.NotifyConf.LineNotifyRepair {
			return
		}
		np := getLevelNum(l.LastLevel)
		if np > nl {
			return
		}
		// send repair notify
		title, message := getLineMessage(l, true)
		if err := SendLine(&datastore.NotifyConf, title+"\n"+message, 0, 0); err != nil {
			log.Printf("send LINE error=%v", err)
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "system",
				Level:    "warn",
				NodeID:   l.NodeID,
				NodeName: l.NodeName,
				Event:    i18n.Trans("Failed to send rapair notify to line"),
			})
			return
		}
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:     "system",
			Level:    "info",
			NodeID:   l.NodeID,
			NodeName: l.NodeName,
			Event:    i18n.Trans("Sent rapair notify to line"),
		})
		return
	}
	np := getLevelNum(l.Level)
	if np > nl {
		return
	}
	title, message := getLineMessage(l, false)
	if err := SendLine(&datastore.NotifyConf, title+"\n"+message, 0, 0); err != nil {
		log.Printf("send discord error=%v", err)
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:     "system",
			Level:    "warn",
			NodeID:   l.NodeID,
			NodeName: l.NodeName,
			Event:    i18n.Trans("Failed to send notify to line"),
		})
		return
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "system",
		Level:    "info",
		NodeID:   l.NodeID,
		NodeName: l.NodeName,
		Event:    i18n.Trans("Sent notify to line"),
	})
}

func getLineMessage(l *datastore.EventLogEnt, repair bool) (string, string) {
	subtitle := i18n.Trans("(Failure)")
	if repair {
		subtitle = i18n.Trans("(Repair)")
	}
	if i18n.GetLang() == "ja" {
		return fmt.Sprintf("%s(%s)", datastore.NotifyConf.Subject, subtitle),
			fmt.Sprintf(
				`発生日時: %s
状態: %s
タイプ: %s
関連ノード: %s
イベント: %s
`,
				formatLogTime(l.Time),
				levelName(l.Level),
				l.Type,
				l.NodeName,
				l.Event,
			)
	}
	return fmt.Sprintf("%s(%s)", datastore.NotifyConf.Subject, subtitle),
		fmt.Sprintf(
			`Time: %s
Status: %s
Type: %s
Node: %s
Event: %s
`,
			formatLogTime(l.Time),
			levelName(l.Level),
			l.Type,
			l.NodeName,
			l.Event,
		)
}
