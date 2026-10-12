import { lang } from '../i18n/i18n';

export interface PollingModeOption {
  name: string;
  value: string;
}

export interface PollingTypeGuide {
  placeholder: string;
  help: string;
  filterHelp?: string;
  scriptHelp?: string;
}

// 判定用: 言語が日本語かどうか
const checkJa = (t?: (k: string) => string): boolean => {
  if (t) {
    const s = t("Polling.Save");
    if (s === "保存") return true;
    if (s === "Save") return false;
  }
  return lang === "ja";
};

// モード定義マップ
export const getPredefinedModes = (type: string, isJa: boolean): PollingModeOption[] | null => {
  switch (type) {
    case "ping":
      return [
        { name: isJa ? "通常 (Ping)" : "Normal (Ping)", value: "" },
        { name: isJa ? "ライン品質測定 (line)" : "Line Quality (line)", value: "line" },
        { name: isJa ? "連続Ping統計 (smoke)" : "Smokeping (smoke)", value: "smoke" },
      ];
    case "snmp":
      return [
        { name: isJa ? "get (OID GET / 値取得)" : "get (OID GET / Value)", value: "get" },
        { name: isJa ? "delta (前回値との差分)" : "delta (Diff from last)", value: "delta" },
        { name: isJa ? "ps (1秒あたりの増分 rate)" : "ps (Rate per sec)", value: "ps" },
        { name: isJa ? "sysUpTime (稼働時間・再起動検知)" : "sysUpTime (Uptime/Reboot)", value: "sysUpTime" },
        { name: isJa ? "ifOperStatus (ポート稼働状態)" : "ifOperStatus (Port Status)", value: "ifOperStatus" },
        { name: isJa ? "traffic (トラフィック bps/pps)" : "traffic (Traffic bps/pps)", value: "traffic" },
        { name: isJa ? "count (Walk マッチ行数カウント)" : "count (Walk Match Count)", value: "count" },
        { name: isJa ? "process (プロセス一覧・変化検知)" : "process (Process Monitor)", value: "process" },
        { name: isJa ? "stats (テーブル数値集計 sum/avg)" : "stats (Table Stats sum/avg)", value: "stats" },
        { name: isJa ? "hrSystemDate (システム時刻差分)" : "hrSystemDate (System Clock Diff)", value: "hrSystemDate" },
        { name: isJa ? "script (snmpGet スクリプト)" : "script (snmpGet Script)", value: "script" },
      ];
    case "http":
      return [
        { name: isJa ? "通常 (HTTP GET)" : "Normal (HTTP GET)", value: "" },
        { name: isJa ? "https (証明書検証あり)" : "https (Strict TLS verify)", value: "https" },
        { name: isJa ? "hash (SHA256 改ざん検知)" : "hash (SHA256 Hash check)", value: "hash" },
        { name: isJa ? "metrics (サーバーメトリクス解析)" : "metrics (Server metrics)", value: "metrics" },
        { name: isJa ? "metrics/json (JSONメトリクス)" : "metrics/json (JSON metrics)", value: "metrics/json" },
      ];
    case "tcp":
      return [
        { name: isJa ? "通常 (ポート接続確認)" : "Normal (Port connect)", value: "" },
      ];
    case "tls":
      return [
        { name: isJa ? "verify (証明書検証)" : "verify (Verify cert)", value: "verify" },
        { name: isJa ? "cert (サーバー証明書レポート)" : "cert (Certificate report)", value: "cert" },
        { name: isJa ? "version (TLSバージョン確認)" : "version (TLS version check)", value: "version" },
        { name: isJa ? "expire (有効期限確認)" : "expire (Expiration check)", value: "expire" },
      ];
    case "dns":
      return [
        { name: isJa ? "ipaddr (A/AAAA・IP変化検知)" : "ipaddr (A/AAAA IP change)", value: "ipaddr" },
        { name: isJa ? "addr (正引き A/AAAA)" : "addr (Forward A/AAAA)", value: "addr" },
        { name: isJa ? "host (逆引き PTR)" : "host (Reverse PTR)", value: "host" },
        { name: isJa ? "mx (メールサーバー MX)" : "mx (Mail Server MX)", value: "mx" },
        { name: isJa ? "ns (ネームサーバー NS)" : "ns (Name Server NS)", value: "ns" },
        { name: isJa ? "txt (テキストレコード TXT)" : "txt (Text Record TXT)", value: "txt" },
        { name: isJa ? "cname (エイリアス CNAME)" : "cname (Alias CNAME)", value: "cname" },
      ];
    case "stun":
      return [
        { name: isJa ? "udp4 (IPv4)" : "udp4 (IPv4)", value: "" },
        { name: isJa ? "ipv6 / udp6 (IPv6)" : "ipv6 / udp6 (IPv6)", value: "ipv6" },
      ];
    case "syslog":
      return [
        { name: isJa ? "count (ログ件数カウント)" : "count (Log count)", value: "count" },
        { name: isJa ? "pri (プライオリティ別集計)" : "pri (Priority stats)", value: "pri" },
        { name: isJa ? "stats (統計集計)" : "stats (Aggregation)", value: "stats" },
        { name: isJa ? "sigma (Sigma脅威検知ルール)" : "sigma (Sigma threat rules)", value: "sigma" },
      ];
    case "trap":
      return [
        { name: isJa ? "count (TRAP件数カウント)" : "count (TRAP count)", value: "count" },
        { name: isJa ? "stats (統計集計)" : "stats (Aggregation)", value: "stats" },
      ];
    case "arplog":
      return [
        { name: isJa ? "count (ARPイベントカウント)" : "count (ARP event count)", value: "count" },
        { name: isJa ? "stats (統計集計)" : "stats (Aggregation)", value: "stats" },
      ];
    case "netflow":
      return [
        { name: isJa ? "traffic (流量 bps/pps)" : "traffic (Traffic bps/pps)", value: "traffic" },
        { name: isJa ? "count (フロー件数)" : "count (Flow count)", value: "count" },
        { name: isJa ? "stats (統計集計)" : "stats (Aggregation)", value: "stats" },
      ];
    case "gnmi":
      return [
        { name: isJa ? "get (テレメトリ取得)" : "get (Telemetry get)", value: "get" },
        { name: isJa ? "subscribe (購読ストリーム)" : "subscribe (Stream subscribe)", value: "subscribe" },
      ];
    case "mqtt":
      return [
        { name: isJa ? "subscribe (トピック購読)" : "subscribe (Topic subscribe)", value: "subscribe" },
        { name: isJa ? "connect (接続確認・RTT)" : "connect (Broker connect & RTT)", value: "connect" },
      ];
    case "email":
      return [
        { name: isJa ? "stats (メール件数・サイズ)" : "stats (Mail count & size)", value: "stats" },
        { name: isJa ? "login (接続確認・RTT)" : "login (Login connect & RTT)", value: "login" },
      ];
    case "twlogeye":
      return [
        { name: isJa ? "notify (通知監視)" : "notify (Notification monitor)", value: "notify" },
        { name: "report.syslog", value: "report.syslog" },
        { name: "report.trap", value: "report.trap" },
        { name: "report.netflow", value: "report.netflow" },
        { name: "report.winevent", value: "report.winevent" },
        { name: "report.mqtt", value: "report.mqtt" },
        { name: "report.otel", value: "report.otel" },
        { name: "report.anomaly", value: "report.anomaly" },
      ];
    default:
      return null;
  }
};

// 対象種別が事前定義モードを持つかどうか
export const hasPredefinedModes = (type: string): boolean => {
  return getPredefinedModes(type, false) !== null;
};

// UI用モード選択肢リスト取得（カスタムな値がある場合は選択肢に含める）
export const getModeList = (type: string, currentMode?: string, t?: (k: string) => string): PollingModeOption[] => {
  const isJa = checkJa(t);
  const baseModes = getPredefinedModes(type, isJa);
  if (!baseModes) {
    return [];
  }
  const modes = [...baseModes];

  // snmp などの場合、空文字と "get" は等価なデフォルト扱い
  const modeVal = currentMode ?? "";
  const exists = modes.some((m) => m.value === modeVal || (m.value === "" && modeVal === "get") || (m.value === "get" && modeVal === ""));
  if (!exists && modeVal !== "") {
    // カスタムなモード文字列がある場合は先頭に追加（デフォルトで選ばれるようにする）
    modes.unshift({
      name: `${modeVal} (${isJa ? "カスタム" : "Custom"})`,
      value: modeVal,
    });
  }
  return modes;
};

// パラメータ・フィルター・スクリプトの案内テキスト取得
export const getPollingGuide = (type: string, mode: string, t?: (k: string) => string): PollingTypeGuide => {
  const isJa = checkJa(t);
  const m = mode ?? "";

  switch (type) {
    case "ping":
      if (m === "smoke") {
        return {
          placeholder: "count=10,size=64,ttl=64",
          help: isJa
            ? "count: 連続送信回数, size: パケットサイズ, ttl: TTL (生存時間)"
            : "count: packet count, size: packet bytes, ttl: TTL",
          scriptHelp: isJa
            ? "判定スクリプト例: loss < 10.0 && avg < 100 * 1000 * 1000 (変数: avg, min, max, median, jitter, loss, fail, count)"
            : "Script example: loss < 10.0 && avg < 100 * 1000 * 1000 (vars: avg, min, max, median, jitter, loss, fail, count)",
        };
      }
      return {
        placeholder: isJa ? "64 または size=64,ttl=64" : "64 or size=64,ttl=64",
        help: isJa ? "パケットサイズ (バイト数) または size=64,ttl=64" : "Packet size in bytes, or size=64,ttl=64",
      };

    case "snmp":
      if (m === "delta" || m === "ps") {
        return {
          placeholder: "ifInOctets.$i",
          help: isJa
            ? "数値OIDまたはMIB名 (前回値との差分または1秒あたりの増分を算出)"
            : "Numeric OID or MIB name (calculates delta or rate per second)",
        };
      }
      if (m === "ifOperStatus" || m === "traffic") {
        return {
          placeholder: isJa ? "$i または 1" : "$i or 1",
          help: isJa
            ? "対象ポートのインターフェイスインデックス (ifIndex: $i または 1, 2...)"
            : "Interface index (ifIndex: $i or 1, 2...)",
        };
      }
      if (m === "sysUpTime" || m === "hrSystemDate" || m === "script") {
        return {
          placeholder: isJa ? "(パラメータ指定不要)" : "(No parameter required)",
          help: isJa
            ? m === "script"
              ? "スクリプト欄で snmpGet() を使用して判定します"
              : "パラメータ指定は不要です (自動取得)"
            : m === "script"
              ? "Evaluate with snmpGet() in script"
              : "No parameter required (automatically gathered)",
        };
      }
      if (m === "count") {
        return {
          placeholder: "hrSWRunName",
          help: isJa
            ? "Walk対象のMIBテーブル名またはOID"
            : "MIB table name or OID to walk",
          filterHelp: isJa
            ? "フィルター欄にマッチさせる条件 (正規表現) を指定"
            : "Specify regex condition to match in Filter",
          scriptHelp: isJa ? "Script例: count > 0" : "Script example: count > 0",
        };
      }
      if (m === "process") {
        return {
          placeholder: isJa ? "(パラメータ指定不要)" : "(No parameter required)",
          help: isJa
            ? "パラメータ指定不要 (フィルター欄に監視対象プロセス名を指定)"
            : "No parameter needed (specify target process name in Filter)",
          filterHelp: isJa
            ? "監視対象のプロセス名 (正規表現)"
            : "Target process name (regex)",
          scriptHelp: isJa ? "Script例: changed == 0 (PID変化検知)" : "Script example: changed == 0 (PID change check)",
        };
      }
      if (m === "stats") {
        return {
          placeholder: "hrProcessorLoad",
          help: isJa
            ? "集計対象のMIBテーブル項目名 (hrProcessorLoad, hrStorageUsedなど)"
            : "MIB table item name to aggregate (hrProcessorLoad, hrStorageUsed, etc.)",
          scriptHelp: isJa ? "Script例: avg < 80.0 && max < 95.0" : "Script example: avg < 80.0 && max < 95.0",
        };
      }
      return {
        placeholder: "sysUpTime.0, hrStorageSize.$i",
        help: isJa
          ? "取得するOIDまたはMIB名 (カンマ区切りで複数指定可)"
          : "OID or MIB name to GET (comma-separated for multiple)",
      };

    case "http":
      if (m === "metrics" || m === "metrics/json") {
        return {
          placeholder: "http://192.168.1.1:9100/metrics",
          help: isJa
            ? "メトリクスエンドポイントURL (Prometheus, Nginx, Apache等)"
            : "Metrics endpoint URL (Prometheus, Nginx, Apache, etc.)",
        };
      }
      if (m === "hash") {
        return {
          placeholder: "http://192.168.1.1/",
          help: isJa
            ? "改ざん・更新検知対象のURL (コンテンツSHA256ハッシュの変化を検知)"
            : "Target URL to monitor content SHA256 hash changes",
        };
      }
      if (m === "https") {
        return {
          placeholder: "https://192.168.1.1/",
          help: isJa ? "HTTPS URL (証明書を厳格に検証)" : "HTTPS URL (strictly verifies TLS certificate)",
        };
      }
      return {
        placeholder: isJa ? "http://192.168.1.1/ または /health または 80" : "http://192.168.1.1/ or /health or 80",
        help: isJa ? "URL、URLパス、またはポート番号" : "URL, URL path, or port number",
      };

    case "tcp":
      return {
        placeholder: isJa ? "80 または 443" : "80 or 443",
        help: isJa ? "接続確認するポート番号" : "Port number to check connection",
      };

    case "tls":
      if (m === "expire") {
        return {
          placeholder: isJa ? "443 または example.com:443" : "443 or example.com:443",
          help: isJa
            ? "ポート番号 または ホスト名:ポート番号"
            : "Port number or hostname:port",
          scriptHelp: isJa
            ? "Script欄に期限残日数 (例: 30) を指定"
            : "Specify remaining days threshold (e.g. 30) in Script",
        };
      }
      return {
        placeholder: isJa ? "443 または example.com:443" : "443 or example.com:443",
        help: isJa ? "ポート番号 または ホスト名:ポート番号" : "Port number or hostname:port",
      };

    case "dns":
      if (m === "host") {
        return {
          placeholder: "8.8.8.8",
          help: isJa ? "逆引き解決するIPアドレス (PTRレコード)" : "IP address for reverse lookup (PTR)",
        };
      }
      if (m === "mx" || m === "ns" || m === "txt") {
        return {
          placeholder: "example.com",
          help: isJa ? "問い合わせるドメイン名" : "Domain name to query",
        };
      }
      return {
        placeholder: "www.google.com",
        help: isJa
          ? m === "ipaddr"
            ? "監視対象ホスト名 (A/AAAA解決によるIPアドレス変化を検知)"
            : "正引き解決するホスト名"
          : m === "ipaddr"
            ? "Target hostname (detects IP address changes via A/AAAA)"
            : "Hostname to resolve",
      };

    case "stun":
      return {
        placeholder: "stun.cloudflare.com:3478",
        help: isJa
          ? "STUNサーバー (ホスト名:ポート、空欄時はデフォルト: stun.cloudflare.com:3478)"
          : "STUN server (host:port, default: stun.cloudflare.com:3478)",
        scriptHelp: isJa
          ? "利用可能変数: ip, oldip, port, rtt (スクリプト未入力時はIP変化時に障害)"
          : "Available vars: ip, oldip, port, rtt (detects IP change if empty)",
      };

    case "syslog":
      if (m === "sigma") {
        return {
          placeholder: "rules/windows/sysmon_susp_powershell.yml",
          help: isJa
            ? "Sigmaルールファイルパス または ルール識別名"
            : "Sigma rule file path or rule name",
        };
      }
      return {
        placeholder: isJa ? "Host または Tag (空欄で全件)" : "Host or Tag (blank for all)",
        help: isJa
          ? "集計キー (Host, Tag, Priority等。空欄時は全件カウント)"
          : "Aggregation key (Host, Tag, etc. Blank for total count)",
        filterHelp: isJa ? "ログ内容を正規表現で絞り込み" : "Filter logs by regex pattern",
      };

    case "trap":
      return {
        placeholder: isJa ? "From または TrapType (空欄で全件)" : "From or TrapType (blank for all)",
        help: isJa
          ? "TRAP集計キー (From, TrapType等。空欄時は全件カウント)"
          : "TRAP aggregation key (From, TrapType, etc. Blank for total)",
        filterHelp: isJa ? "TRAP内容を正規表現で絞り込み" : "Filter traps by regex pattern",
      };

    case "arplog":
      return {
        placeholder: isJa ? "IP または MAC (空欄で全件)" : "IP or MAC (blank for all)",
        help: isJa ? "ARPイベント集計キー (空欄時は全件カウント)" : "ARP event aggregation key (Blank for total)",
      };

    case "netflow":
      if (m === "stats") {
        return {
          placeholder: "SrcAddr, DstAddr",
          help: isJa ? "NetFlow統計集計キー (SrcAddr, DstAddr, Protocol等)" : "NetFlow aggregation keys (SrcAddr, DstAddr, etc.)",
        };
      }
      return {
        placeholder: isJa ? "(空欄可)" : "(Blank)",
        help: isJa ? "パラメータ空欄 (フィルターでフロー条件を絞り込み可)" : "No parameter needed (filter flows by Filter)",
      };

    case "gnmi":
      return {
        placeholder: "/interfaces/interface[name=*]/state/counters",
        help: isJa ? "gNMI Path (問い合わせまたは購読するテレメトリパス)" : "gNMI Path to query or subscribe",
      };

    case "mqtt":
      if (m === "connect") {
        return {
          placeholder: isJa ? "(パラメータ指定不要)" : "(No parameter required)",
          help: isJa ? "MQTTブローカー接続確認 (パラメータ不要)" : "Broker connection check (no params needed)",
        };
      }
      return {
        placeholder: "twsnmpfk/sensor/#",
        help: isJa ? "購読するMQTTトピック名" : "MQTT topic name to subscribe",
      };

    case "email":
      return {
        placeholder: "imaps://user:pass@mail.example.com:993",
        help: isJa
          ? "URL形式: imap://user:pass@host:port または pop3://user:pass@host:port"
          : "URL format: imap://user:pass@host:port or pop3://user:pass@host:port",
        scriptHelp: isJa
          ? "利用可能変数: count (件数), size (バイト), newMsg (新規判定), lastDate (最終日時)"
          : "Available vars: count, size, newMsg, lastDate",
      };

    case "cmd":
      return {
        placeholder: "ping -c 1 127.0.0.1",
        help: isJa
          ? "ローカルで実行するコマンド (標準出力をFilter/Extractor/Scriptで判定)"
          : "Local command to execute (evaluate stdout with Filter/Extractor/Script)",
      };

    case "ssh":
      return {
        placeholder: "uptime",
        help: isJa
          ? "SSH接続先で実行するリモートコマンド"
          : "Remote command to execute via SSH",
      };

    case "ntp":
      return {
        placeholder: isJa ? "123 (空欄でデフォルト123)" : "123 (default 123 if blank)",
        help: isJa ? "NTPポート番号 (空欄時は標準の123番ポート)" : "NTP port number (default 123)",
      };

    case "pihole":
      return {
        placeholder: "/admin/api.php",
        help: isJa ? "Pi-Hole APIパスまたはトークン" : "Pi-Hole API path or token",
      };

    case "lxi":
      return {
        placeholder: "*IDN?",
        help: isJa ? "LXI計測機器用SCPIコマンド" : "SCPI command for LXI instrument",
      };

    case "report":
      return {
        placeholder: isJa ? "ip または user" : "ip or user",
        help: isJa ? "集計対象レポート名" : "Target report name",
      };

    case "twlogeye":
      return {
        placeholder: isJa ? "(空欄可)" : "(Blank)",
        help: isJa ? "TwLogEye連携設定" : "TwLogEye integration settings",
      };

    default:
      return {
        placeholder: isJa ? "パラメータを入力" : "Enter parameters",
        help: isJa ? "種別に応じたパラメータを指定してください" : "Specify parameters according to type",
      };
  }
};
