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
	"Add Network": {
		"ja": "ネットワークを追加",
	},
	"Update Network": {
		"ja": "ネットワークを更新",
	},
	"Delete Network": {
		"ja": "ネットワークを削除",
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
	"Delete Polling logs(%d)": {
		"ja": "ポーリングログを削除(%d件)",
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
	"Update by dicover": {
		"ja": "自動発見により更新",
	},
	"ARP Watch range %s usage:%d/%d %.2f%%": {
		"ja": "ARP監視 %s アドレス使用量:%d/%d %.2f%%",
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
	"Delete all NetFlow logs": {
		"ja": "NetFlow全削除",
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
	"Delete arp ent ip=%s": {
		"ja": "IPアドレス%sのARP監視情報を削除",
	},
	"Clear all arp watch info": {
		"ja": "ARP監視情報を全削除",
	},
	"Failed to send rapair notify to line": {
		"ja": "復帰通知をLINEへ送信できません",
	},
	"Sent rapair notify to line": {
		"ja": "復帰通知をLINEへ送信しました",
	},
	"Failed to send notify to line": {
		"ja": "障害通知をLINEへ送信できません",
	},
	"Sent notify to line": {
		"ja": "障害通知をLINEへ送信しました",
	},
	"(LINE test)": {
		"ja": "(LINEのテスト)",
	},
	"Confirm init ssh key": {
		"ja": "SSH秘密鍵の初期化",
	},
	"Do you want to init?": {
		"ja": "初期化しますか？",
	},
	"Init ssh private key": {
		"ja": "SSH秘密鍵を初期化しました",
	},
	"Export MAP": {
		"ja": "マップを保存",
	},
	"Low memory alert": {
		"ja": "メモリー不足",
	},
	"Low storage alert": {
		"ja": "ストレージ容量不足",
	},
	"Over load alert": {
		"ja": "高負荷状態",
	},
	"Start PKI": {
		"ja": "PKIを開始しました",
	},
	"Create CA Certificate subject=%s serial=%x": {
		"ja": "CA証明書を作成しました subject=%s serial=%x",
	},
	"Reject CSR subject=%s info=%+v err=%v": {
		"ja": "証明書の発行を却下しました subject=%s info=%+v err=%v",
	},
	"Create Certificate subject=%s serial=%s info=%+v": {
		"ja": "証明書を発行しました subject=%s serial=%s info=%+v",
	},
	"CA is already valid": {
		"ja": "CAはすでに構築済みです",
	},
	"Can not create CA err=%v": {
		"ja": "CAを構築できません err=%v",
	},
	"Create CA": {
		"ja": "CAを構築しました",
	},
	"Confirm destroy CA": {
		"ja": "CAの破棄",
	},
	"Do you want to destroy CA?": {
		"ja": "CAを破棄しますか？",
	},
	"Destroy CA": {
		"ja": "CAを破棄しました",
	},
	"Can not create CSR err=%v": {
		"ja": "証明書要求を作成できません err=%v",
	},
	"Can not create Certificate err=%v": {
		"ja": "証明書を発行できません err=%v",
	},
	"Confirm revoke cert": {
		"ja": "証明書の失効",
	},
	"Do you want to revoke selected Cert?": {
		"ja": "選択した証明書を失効しますか?",
	},
	"Revoke Cert subject=%s serial=%s": {
		"ja": "証明書を失効しました subject=%s serial=%s",
	},
	"Create SCEP CA Certificate subject=%s serial=%x": {
		"ja": "SCEP用のCA証明証を作成しました subject=%s serial=%x",
	},
	"Start CRL/OCSP/SCEP Server port=%d": {
		"ja": "CRL/OCSP/SCEPサーバー起動 port=%d",
	},
	"Stop CRL/OCSP/SCEP Server": {
		"ja": "CRL/OCSP/SCEPサーバー停止",
	},
	"Create ACME Server Certificate subject=%s serial=%x": {
		"ja": "ACMEサーバー用の証明証を発行しました subject=%s serial=%x",
	},
	"Start ACME Server port=%d": {
		"ja": "ACMEサーバー起動 port=%d",
	},
	"Stop ACME Server": {
		"ja": "ACMEサーバー停止",
	},
	"MAC Address": {
		"ja": "MACアドレス",
	},
	"Node": {
		"ja": "ノード名",
	},
	"IP Address": {
		"ja": "IPアドレス",
	},
	"Descr": {
		"ja": "説明",
	},
	"Managed Node": {
		"ja": "管理対象ノード",
	},
	"No": {
		"ja": "いいえ",
	},
	"Vendor": {
		"ja": "ベンダー",
	},
	"ARP Watch": {
		"ja": "ARP監視",
	},
	"Not fond": {
		"ja": "不明",
	},
	"DNS Host": {
		"ja": "DNSホスト名",
	},
	"Location": {
		"ja": "位置",
	},
	"Start OpenTelemetry server": {
		"ja": "OpenTelemetryサーバー受信開始",
	},
	"Stop OpenTelemetry server": {
		"ja": "OpenTelemetryサーバー受信停止",
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
