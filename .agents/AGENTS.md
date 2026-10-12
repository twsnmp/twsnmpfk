# twsnmpfk Project Rules

## 1. Prohibition of `window.confirm()`

Never use JavaScript native `window.confirm()` or `confirm()` in frontend code.

- **Reason**: In Wails v2 (especially on macOS WKWebView / WebKit), native dialog handlers for JavaScript `confirm()` and `alert()` are not implemented by default. Calling `window.confirm()` silently and immediately returns `false` without presenting any prompt to the user, causing actions (such as deletion or confirmation workflows) to fail silently.
- **Proper Implementation**: Use inline confirmation buttons (e.g. switching the action button to `[ Delete ] [ Cancel ]` state within the row or component) or custom Flowbite/Svelte modals, instead of native browser dialogs.

## 2. DataTables Usage inside Flowbite / Svelte Tab Content

When using DataTables (`datatables.net`) inside Flowbite's `<TabItem>` or Svelte slot content, always wrap the `<table>` element with a `<div>`.

- **Reason**: DataTables injects a `<div id="..._wrapper">` around `<table>` during initialization. If `<table>` is a direct child element of a slot or tab panel, this DOM manipulation breaks sibling elements (such as echarts chart containers).
- **Correct Pattern**:
  ```html
  <!-- Wrap in a div to contain DataTables DOM manipulations -->
  <div>
    <table id="myTable" class="display">...</table>
  </div>
  ```
- **Incorrect Pattern**:
  ```html
  <!-- Direct child of Flowbite/Svelte slot -->
  <table id="myTable" class="display">...</table>
  ```
- **Real-world Example**: Occurred in `SyslogReport.svelte` (`#syslogSummaryTable`). When DataTables reinitialized with `destroy: true`, its sibling `<div id="syslogSummary">` (echarts container) was destroyed, causing charts in other tabs to disappear.

## 3. Build, Verification, and Testing Rules

- **Build & Verification**: The user runs full builds and application verification using `task dev`. Agents must not run full builds or application verification.
- **Test Execution**: Agents must only run individual unit tests directly related to the modified code. Do not run full or unrelated test suites.
