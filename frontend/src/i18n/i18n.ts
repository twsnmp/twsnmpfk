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

let initialLang = "ja";
try {
  const hasParam = await HasLangParam().catch(() => false);
  const backendLang = await GetLang().catch(() => "");
  const savedLang = typeof localStorage !== "undefined" ? localStorage.getItem("twsnmpfk_lang") : null;

  if (hasParam) {
    initialLang = backendLang || "ja";
    if (typeof localStorage !== "undefined") {
      localStorage.setItem("twsnmpfk_lang", initialLang);
    }
  } else if (savedLang === "ja" || savedLang === "en") {
    initialLang = savedLang;
    SetLang(savedLang).catch(() => {});
  } else {
    initialLang = backendLang || "ja";
  }
} catch (e) {
  console.warn("i18n initialization fallback:", e);
}

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
  if (typeof localStorage !== "undefined") {
    localStorage.setItem("twsnmpfk_lang", newLang);
  }
  await SetLang(newLang).catch((err) => {
    console.error("SetLang error:", err);
  });
};
