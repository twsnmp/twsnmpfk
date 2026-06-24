# twsnmpfk Project Rules

## DataTables を Flowbite / Svelte タブコンテンツ内で使用する場合のルール

DataTables（`datatables.net`）を Flowbite の `<TabItem>` や Svelte のスロットコンテンツ内で使用する際は、`<table>` 要素を必ず `<div>` で囲むこと。

**理由**: DataTables は初期化時に `<table>` の周囲に `<div id="..._wrapper">` を注入する。`<table>` がスロット・タブパネルの直接子要素だと、この DOM 操作が兄弟要素（echarts コンテナ等）を破壊する。

**正しい実装**:
```html
<!-- ✅ 正しい: div でラップして DataTables の DOM 操作を封じ込める -->
<div>
  <table id="myTable" class="display">...</table>
</div>
```

**間違った実装**:
```html
<!-- ❌ 間違い: Flowbite/Svelte スロットの直接子要素として <table> を置く -->
<table id="myTable" class="display">...</table>
```

**実例**: `SyslogReport.svelte` の `#syslogSummaryTable` で発生。DataTables の `destroy: true` 再初期化時に兄弟の `<div id="syslogSummary">` (echarts コンテナ) が削除され、他タブのグラフが表示されなくなった。
