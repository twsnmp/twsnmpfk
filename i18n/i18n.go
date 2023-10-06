package i18n

import (
	"log"
	"strings"

	"github.com/Xuanwo/go-locale"
)

var lang = ""

// 翻訳マップ
var transMap = map[string]map[string]string{
	"Start TWSNMP": {
		"ja": "TWSNMP起動",
	},
	"Stop TWSNMP": {
		"ja": "TWSNMP停止",
	},
	"Delete AI Result(%s)": {
		"ja": "AI分析結果を削除(%s)",
	},
	"Add line to %s": {
		"ja": "ラインを%sに接続",
	},
	"Update line to %s": {
		"ja": "%sとのラインを更新",
	},
	"Add draw item": {
		"ja": "描画アイテムを追加",
	},
	"Update draw item": {
		"ja": "描画アイテムを更新",
	},
	"Delete draw item(%d)": {
		"ja": "描画アイテムを削除(%d件)",
	},
	"Copy draw item": {
		"ja": "描画アイテムをコピー",
	},
	"Add Node": {
		"ja": "ノードを追加",
	},
	"Update Node": {
		"ja": "ノードを更新",
	},
	"Delete Node": {
		"ja": "ノードを削除",
	},
	"Copy Node": {
		"ja": "ノードをコピー",
	},
	"Send Wake on LAN Packet to %s": {
		"ja": "Wake on LANパケットを%sに送信",
	},
	"Delete Polling(%d)": {
		"ja": "ポーリングを削除(%d件)",
	},
	"AI report:%s(%s):%f": {
		"ja": "AI分析レポート:%s(%s):%f",
	},
	"node=%d,down=%d,rate=%.2f%%": {
		"ja": "ノード数=%d,障害ノード=%d,稼働率=%.2f%%",
	},
	"TWSNMP verison this=%s latest=%s": {
		"ja": "このTWSNMPのバージョンは%s、最新版は%s",
	},
	"%.2fDays(%d)": {
		"ja": "%.2f日(%d)",
	},
	"%.2fHours(%d)": {
		"ja": "%.2f時間(%d)",
	},
	"%.2fSec(%d)": {
		"ja": "%.2f秒(%d)",
	},
	"Notify from TWSNMP": {
		"ja": "TWSNMPからの通知",
	},
	"Start discover %s - %s": {
		"ja": "自動発見開始 %s - %s",
	},
	"End discover %s - %s": {
		"ja": "自動発見終了 %s - %s",
	},
	"Found at %s": {
		"ja": "%sに発見",
	},
	"Protocol:": {
		"ja": "対応プロトコル:",
	},
	"Add by dicover": {
		"ja": "自動発見により追加",
	},
	"ARP Watch local address usage:%d/%d %.2f%%": {
		"ja": "ARP監視 ローカルアドレス使用量:%d/%d %.2f%%",
	},
	"Change MAC Address %s -> %s": {
		"ja": "MACアドレス変化 %s -> %s",
	},
	"Set MAC Address %s": {
		"ja": "MACアドレス設定 %s",
	},
	"Fixed MAC address node MAC is %s": {
		"ja": "MACアドレス固定ノードのアドレス取得 %s",
	},
	"Fixed MAC address node '%s' Chnage IP address from '%s' to '%s'": {
		"ja": "MACアドレス固定ノード'%s'のIPアドレスが'%s'から'%s'に変化",
	},
	"Fixed host name node '%s' Chnage IP from '%s' to ''%s": {
		"ja": "ホスト名固定ノード'%s'のIPアドレスが'%s'から''%sに変化",
	},
	"Send notify mail %s": {
		"ja": "通知メール送信 %s",
	},
	"Send repair mail %s": {
		"ja": "復帰通知メール送信 %s",
	},
	"(test mail)": {
		"ja": "(試験メール）",
	},
	"Notify command レベル=%d %s": {
		"ja": "外部通知コマンド実行 レベル=%d %s",
	},
	"(Failure)": {
		"ja": "(障害)",
	},
	"(Repair)": {
		"ja": "(復帰)",
	},
	"High": {
		"ja": "重度",
	},
	"Low": {
		"ja": "軽度",
	},
	"Warnning": {
		"ja": "注意",
	},
	"Normal": {
		"ja": "正常",
	},
	"Repair": {
		"ja": "復帰",
	},
	"Unknown": {
		"ja": "不明",
	},
	"High=%d,Low=%d,Warn=%d,Normal=%d,Other=%d": {
		"ja": "重度=%d,軽度=%d,注意=%d,正常=%d,その他=%d",
	},
	"MAP=%s": {
		"ja": "マップ名=%s",
	},
	"MAP State=%s": {
		"ja": "マップ状態=%s",
	},
	"Min:%s%% Avg:%s%% Max:%s%%": {
		"ja": "最小:%s%% 平均:%s%% 最大:%s%%",
	},
	"DB Size": {
		"ja": "DBサイズ",
	},
	"MAP Name": {
		"ja": "マップ名",
	},
	"MAP State": {
		"ja": "マップの状態",
	},
	"Node count by state": {
		"ja": "状態別のノード数",
	},
	"CPU Usage": {
		"ja": "CPU使用率",
	},
	"Memory Usage": {
		"ja": "メモリ使用率",
	},
	"Disk Usage": {
		"ja": "ディスク使用率",
	},
	"System Load": {
		"ja": "システム負荷",
	},
	"Log count by level": {
		"ja": "状態別のログ数",
	},
	"%s(report) at %s": {
		"ja": "%s(定期レポート) at %s",
	},
	"Failed to send report mail err=%v": {
		"ja": "定期レポートメール送信失敗 err=%v",
	},
	"Send report mail": {
		"ja": "定期レポートメール送信",
	},
	"re check polling:": {
		"ja": "ポーリング再確認:",
	},
	"Change polling state:%s(%s)": {
		"ja": "ポーリング状態変化:%s(%s)",
	},
	"Select data store path": {
		"ja": "データストアのパスを選択",
	},
	"Start syslogd": {
		"ja": "Syslog受信開始",
	},
	"Stop syslogd": {
		"ja": "Syslog受信停止",
	},
	"Start snmptrapd": {
		"ja": "SNMP TRAP受信開始",
	},
	"Stop snmptrapd": {
		"ja": "SNMP TRAP受信停止",
	},
	"Start ARP watch": {
		"ja": "ARP監視開始",
	},
	"Stop ARP watch": {
		"ja": "ARP監視停止",
	},
	"Confirm clear": {
		"ja": "クリア確認",
	},
	"Do you want to clear?": {
		"ja": "クリアしますか?",
	},
	"Clear ARP": {
		"ja": "ARP監視情報をクリア",
	},
	"Confirm delete": {
		"ja": "削除確認",
	},
	"Do you want to delete?": {
		"ja": "削除しますか?",
	},
	"Delete all event logs": {
		"ja": "イベントログ全削除",
	},
	"Delete all syslog": {
		"ja": "Syslog全削除",
	},
	"Delete all TRAP logs": {
		"ja": "TRAP全削除",
	},
	"Delete all polling logs": {
		"ja": "ポーリングログ全削除",
	},
	"Backup file": {
		"ja": "バックアップ ファイル",
	},
	"Backup done": {
		"ja": "バックアップ完了",
	},
}

func init() {
	t, err := locale.Detect()
	if err != nil {
		log.Println(err)
	}
	l := t.String()
	a := strings.SplitN(l, "-", 2)
	if len(a) == 2 {
		lang = a[0]
	}
	log.Printf("i18n init lang=%s", lang)
}

// SetLangは言語を設定します。
func SetLang(l string) {
	lang = l
}

// GetLangは言語の設定を返します。
func GetLang() string {
	return lang
}

// Transは翻訳した文字列を返します。
func Trans(s string) string {
	if m, ok := transMap[s]; ok {
		if t, ok := m[lang]; ok {
			return t
		}
	}
	return s
}
