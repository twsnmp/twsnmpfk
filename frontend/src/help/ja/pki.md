# PKI (CA構築前)

認証局（CA）を新規に構築するための初期設定画面。

## 初期構築用パラメータ

* **名前**
  認証局（CA）の登録名称。CA自己署名証明書のSubjectに適用される。
* **DNS名**
  発行する証明書のCDP（CRL配布点）、OCSPアドレス、およびACMEサーバー証明書のSANs（Subject Alternative Names）に含めるホスト名やIPアドレス（カンマ区切りで複数指定可能）。
* **ACMEサーバー基本URL**
  ACMEサーバー機能のベースURL。空欄の場合はホスト名から自動設定される。
* **CRL/OCSP/SCEPサーバー基本URL**
  CRL、OCSP、およびSCEPサーバー機能のベースURL。空欄の場合はホスト名から自動設定される。
* **CAの鍵タイプ**
  CAの秘密鍵・公開鍵ペアの暗号化方式と鍵長（RSA、ECDSAなど）。
* **CA証明書の期間**
  CA証明書の有効期間（年単位）。
* **CRLの更新間隔**
  失効リスト（CRL）を更新する時間周期（時間単位）。
* **証明書の期間**
  CAから発行する各クライアント証明書の有効期限（時間単位）。
* **CRL/OCSP/SCEPサーバーのポート番号**
  CRL/OCSP/SCEP待受HTTPサーバーのポート番号。構築後の変更は不可能。
* **ACMEサーバーのポート番号**
  ACMEサーバーの待受ポート番号。構築後の変更は不可能。

---

# PKI (CA構築後)

CA構築完了後に、発行済みデジタル証明書の管理や手動発行操作を行う画面。

## 発行済み証明書テーブルの項目

* **状態**
  証明書の現在のステータス（有効、期限切れ、失効）。
* **Type**
  証明書の種類（クライアント、サーバーなど）。
* **ID**
  証明書のシリアル番号。
* **Subject**
  証明書の所有者情報（Subject）。
* **関連ノード**
  証明書を取得したデバイス/ノードの名前。
* **開始**
  証明書有効期間の開始日時。
* **終了**
  証明書有効期間の満了日時。
* **失効**
  証明書を失効させた日時。

## ボタンの説明

* **[CSR作成]** ： 新規証明書署名要求（CSR）を作成するダイアログを表示するボタン。
* **[証明書作成]** ： 作成・アップロードされたCSRを読み込んで証明書を署名・発行するボタン。
* **[CA初期化]** ： 現在のCA（認証局データ）をすべて破棄し、未構築状態に戻すボタン。
* **[サーバー制御]** ： PKI関連サービス（ACMEサーバー等）の有効/無効切り替えなどの制御画面を開くボタン。
* **[失効]** ： 選択した証明書を失効（Revoke）させ失効リストへ追加するボタン。
* **[エクスポート]** ： 選択した証明書をファイルとして保存するボタン。
* **[更新]** ： 証明書リストを最新状態に更新するボタン。

## CSR作成画面のパラメータ

* **鍵タイプ**
  CSR生成時の秘密鍵アルゴリズム。
* **名前**
  証明書のCommon Name (CN)。
* **DNS名**
  Subject Alt Nameに含めるDNS名（カンマ区切り）。
* **組織名**
  Organization (O) 名（省略可能）。
* **組織単位**
  Organizational Unit (OU) 名（省略可能）。
* **国コード**
  Country (C) コード（例: JP、省略可能）。
* **州/都道府県名**
  State or Province (ST) 名（省略可能）。
* **市町村名**
  Locality (L) 名（省略可能）。

## サーバー制御画面のパラメータ

* **ACME/ESTサーバー**
  ACME（自動証明書管理環境）および EST（RFC 7030 / Enrollment over Secure Transport）サーバー機能の有効化。
* **CRL/OCSP/SCEPサーバー**
  証明書失効リスト（CRL）、OCSP、およびSCEP配信サーバー機能の有効化。
* **ACMEサーバーの基本URL**
  ACMEおよびESTが応答するベースURL。
* **CRLの更新間隔**
  失効リストの更新間隔（時間単位）。
* **証明書の期間**
  発行する証明書の有効期間（時間単位）。

## ACMEサーバーの対応チャレンジと端末構成証明（Device Attestation）

ACMEサーバーは以下の識別子（Identifier）と検証方式（Challenge）に対応しています。証明書を発行するには、要求元の対象ホストまたはデバイスがTWSNMP FKのマップ上にノードとして登録されている必要があります。

* **DNS識別子 (`dns`)**
  * `http-01` / `tls-alpn-01` / `dns-01` に対応。
  * ワイルドカード証明書（`*.example.com`）の `dns-01` 検証およびノード照合に対応。
* **IPアドレス識別子 (`ip` / RFC 8738)**
  * IPv4 / IPv6 アドレスに対する `http-01` および `tls-alpn-01`（`.in-addr.arpa.` / `.ip6.arpa.`）検証に対応。
* **端末構成証明 (`permanent-identifier` / `hardware-module`、`device-attest-01`)**
  * Apple Managed Device Attestation（`fmt: "apple"`、iOS 16 / iPadOS 16 / macOS 14以降）および YubiKey PIV / Smallstep（`fmt: "step"`）の構成証明検証に対応。
  * 発行される証明書のSAN（Subject Alternative Name）拡張には RFC 4043 `PermanentIdentifier`（OID `1.3.6.1.5.5.7.8.3`）が記録されます。
  * **ノード登録方法** ： デバイスのシリアル番号またはUDID（`PermanentIdentifier`）を、対象ノードの「名前」「説明」「公開鍵」または「メモ」のいずれかに記載しておくことでノードと紐付けられます。

## EST（RFC 7030）サーバー機能

ACMEサーバーが有効な場合、同一のHTTPSポート上でRFC 7030準拠のESTサービスが自動的に提供されます。

* **対応エンドポイント（標準および任意ラベルに対応）**:
  * `GET /.well-known/est/cacerts` または `GET /.well-known/est/:label/cacerts`
    * Root CA証明書を含むPKCS#7 (`certs-only`) をBase64形式で返却します。
  * `POST /.well-known/est/simpleenroll` または `POST /.well-known/est/:label/simpleenroll`
    * PKCS#10 CSRを受け取り、認証およびノード照合後に署名済み証明書をPKCS#7 Base64形式で返却します。
  * `POST /.well-known/est/simplereenroll` または `POST /.well-known/est/:label/simplereenroll`
    * 既存証明書の更新を行います。
* **認証方式（mTLS または HTTP Basic認証が必須）**:
  * **mTLS（クライアント証明書認証）**: 当CAから発行された有効かつ未失効のクライアント証明書を用いたTLS接続の場合、認証が成功し証明書が発行・更新されます。
  * **HTTP Basic認証**: クライアント証明書がない場合、CSRが対応する対象ノードのパスワード（`node.Password`）およびユーザー名（`node.User` が設定されている場合）によるHTTP Basic認証が必須となります（未認証の場合は `401 Unauthorized` が返されます）。
* **ノード照合**:
  * CSR内のSAN（DNS名、IPアドレス、RFC 4043 PermanentIdentifier）、Subject CN、またはアクセス元IPが登録ノードと一致している必要があります。

