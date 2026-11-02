import { addMessages, init, locale } from "svelte-i18n";
import { writable } from "svelte/store";
import en from "./en.json";
import ja from "./ja.json";
import {
  GetLang,
  HasLangParam,
  SetLang,
} from "../../wailsjs/go/main/App";

addMessages("en", en);
addMessages("ja", ja);

// ブラウザ側の OS 言語判定（フォールバック用）
const getBrowserLang = () => {
  if (typeof navigator !== "undefined" && navigator.language) {
    return navigator.language.toLowerCase().startsWith("ja") ? "ja" : "en";
  }
  return "en";
};

let initialLang = "en";
try {
  const hasParam = await HasLangParam().catch(() => false);
  const backendLang = await GetLang().catch(() => "");

  if (hasParam && backendLang) {
    // 起動パラメータ(-lang)またはiniで指定された言語
    initialLang = backendLang;
  } else if (backendLang === "ja" || backendLang === "en") {
    // バックエンドで検出したOSの言語設定
    initialLang = backendLang;
  } else {
    // フォールバック: ブラウザのnavigator.languageから検出
    initialLang = getBrowserLang();
  }
} catch (e) {
  console.warn("i18n initialization fallback:", e);
  initialLang = getBrowserLang();
}

// 念のため以前の古いキャッシュキーがあれば削除
try {
  if (typeof localStorage !== "undefined") {
    localStorage.removeItem("twsnmpfk_lang");
  }
} catch {}

export let lang = initialLang;
export const currentLang = writable<string>(initialLang);

init({
  fallbackLocale: "en",
  initialLocale: initialLang,
});

/**
 * 言語を即時リアクティブに切り替えて永続化する（画面リロード不要）
 */
export const changeLang = async (newLang: string) => {
  if (newLang !== "ja" && newLang !== "en") return;
  lang = newLang;
  currentLang.set(newLang);
  locale.set(newLang);
  await SetLang(newLang).catch((err) => {
    console.error("SetLang error:", err);
  });
};
