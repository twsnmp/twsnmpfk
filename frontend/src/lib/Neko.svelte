<script lang="ts">
  import { Modal } from "flowbite-svelte";
  import { onDestroy } from "svelte";
  
  // すべての画像をインポート
  import neko_ng from "../assets/images/neko_ng.png";
  import neko_ok from "../assets/images/neko_ok.png";
  import neko1 from "../assets/images/neko_anm1.png";
  import neko2 from "../assets/images/neko_anm2.png";
  import neko3 from "../assets/images/neko_anm3.png";
  import neko4 from "../assets/images/neko_anm4.png";
  import neko5 from "../assets/images/neko_anm5.png";
  import neko6 from "../assets/images/neko_anm6.png";
  import neko7 from "../assets/images/neko_anm7.png";

  // 外部から制御するプロパティ
  export let show: boolean = false;
  export let status: "" | "waiting" | "ok" | "ng" = "";

  const nekos = [neko1, neko2, neko3, neko4, neko5, neko6, neko7];
  let nekoNo = 0;
  let timer: any = undefined;

  // 表示される画像を計算
  $: currentImg = (() => {
    if (status === "ok") return neko_ok;
    if (status === "ng") return neko_ng;
    return nekos[nekoNo]; // waiting の時
  })();

  // アニメーション制御
  $: if (show && status === "waiting") {
    if (!timer) startAnimation();
  } else {
    stopAnimation();
  }

  function startAnimation() {
    timer = setInterval(() => {
      nekoNo = (nekoNo + 1) % nekos.length;
    }, 200);
  }

  function stopAnimation() {
    clearInterval(timer);
    timer = undefined;
  }

  onDestroy(stopAnimation);
</script>

<Modal
  bind:open={show}
  size="sm"
  dismissable={false}
  class="w-full bg-white bg-opacity-75 dark:bg-white"
>
  <div class="flex justify-center items-center">
    <img src={currentImg} alt="neko status" />
  </div>
</Modal>
