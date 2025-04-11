import P5 from "p5";
import { getIconCode, getStateColor } from "./common";
import {
  GetSettings,
  GetNodes,
  GetLines,
  GetDrawItems,
  GetBackImage,
  UpdateDrawItemPos,
  UpdateDrawItem,
  UpdateNodePos,
  GetImage,
  GetMapConf,
  GetNotifyConf,
  GetNetworks,
  UpdateNetworkPos,
  GetImageIcon,
} from "../../wailsjs/go/main/App";
import type { datastore } from "wailsjs/go/models";
import { gauge, line, bar } from "./chart/drawitem";
import port from "../assets/images/port.png";

let mapSizeX = window.screen.width > 4000 ? 5000 : 2500;
let mapSizeY = 5000;
let mapRedraw = true;
let readOnly = false;
let settingsLock = false;

let mapCallBack: any = undefined;

let nodes: any = {};
let lines: any = [];
let items: any = {};
let networks: any = {};
let backImage: datastore.BackImageEnt = {
  X: 0,
  Y: 0,
  Width: 0,
  Height: 0,
  Path: "",
};

let _backImage: any = undefined;

let fontSize = 12;
let iconSize = 32;

const selectedNodes: any = [];
const selectedDrawItems: any = [];
let selectedNetwork = "";

const imageMap = new Map();
let mapState = 0;
let showAllItems = false;

let _mapP5: P5 | undefined = undefined;
let beepHigh: any = undefined;
let beepLow: any = undefined;
let scale = 1.0;
let mapConf: any = undefined;
let portImage: any = undefined;

export const initMAP = async (div: HTMLElement, cb: any) => {
  const settings = await GetSettings();
  const notifyConf = await GetNotifyConf();
  mapConf = await GetMapConf();
  switch (mapConf.MapSize) {
    case 1:
      mapSizeX = 2894;
      mapSizeY = 4093;
      break;
    case 2:
      mapSizeX = 4093;
      mapSizeY = 2894;
      break;
    default:
      // Auto
      mapSizeX = window.screen.width > 4000 ? 5000 : 2500;
      mapSizeX = 5000;
  }
  beepHigh = notifyConf.BeepHigh;
  beepLow = notifyConf.BeepLow;

  mapCallBack = cb;
  settingsLock = settings.Lock != "";
  readOnly = settingsLock;
  mapRedraw = false;
  if (_mapP5 != undefined) {
    return;
  }
  div.oncontextmenu = (e) => {
    e.preventDefault();
  };
  _mapP5 = new P5(mapMain, div);
};

let lastBackImagePath = "";

export const updateMAP = async () => {
  if (!_mapP5) {
    return;
  }
  const dark = isDark();
  if (!mapConf) {
    mapConf = await GetMapConf();
  }
  const z = mapConf.IconSize || 3;
  iconSize = 8 + z * 8;
  fontSize = 6 + z * 2;
  nodes = await GetNodes();
  lines = await GetLines();
  items = (await GetDrawItems()) || {};
  networks = (await GetNetworks()) || {};
  backImage = await GetBackImage();
  if (!portImage) {
    portImage = _mapP5.loadImage(port);
  }
  if (backImage.Path != lastBackImagePath) {
    if (backImage.Path) {
      const img = _mapP5.loadImage(await GetImage(backImage.Path));
      if (img) {
        _backImage = img;
      }
    } else {
      _backImage = null;
    }
    lastBackImagePath = backImage.Path;
  }
  _setMapState();
  _checkBeep();
  const backColor = dark
    ? _mapP5.color(23).toString()
    : _mapP5.color(252).toString();
  for (const k in nodes) {
    const icon = nodes[k].Image;
    if (icon && !imageMap.has(icon)) {
      const img = _mapP5.loadImage(await GetImageIcon(icon));
      if (img) {
        imageMap.set(icon, img);
      }
    }
  }
  for (const k in items) {
    switch (items[k].Type) {
      case 3:
        if (!imageMap.has(items[k].ID)) {
          const img = _mapP5.loadImage(await GetImage(items[k].Path));
          if (img) {
            imageMap.set(items[k].ID, img);
          }
        }
        break;
      case 2:
      case 4:
        if (items[k].Text.length == 0) {
          items[k].Text = "Empty Draw Item";
        }
        items[k].W = items[k].Size * items[k].Text.length;
        items[k].H = items[k].Size;
        if (!dark) {
          items[k].Color = items[k].Color != "#eee" ? items[k].Color : "#333";
        }
        break;
      case 5:
        items[k].H = items[k].Size * 10;
        items[k].W = items[k].Size * 10;
        if (items[k].Value < 0.001) {
          items[k].Value = 0.0;
        }
        break;
      case 6: { // New Gauge
        items[k].W = items[k].H;
        const img = _mapP5.loadImage(
          gauge(items[k].Text || "", items[k].Value || 0, backColor)
        );
        if (img) {
          imageMap.set(k, img);
        }
        break;
      }
      case 7: { // Bar
        items[k].W = items[k].H * 4;
        const img = _mapP5.loadImage(
          bar(
            items[k].Text || "",
            items[k].Color || "white",
            items[k].Value || 0,
            backColor
          )
        );
        if (img) {
          imageMap.set(k, img);
        }
        break;
      }
      case 8: { // Line
        items[k].W = items[k].H * 4;
        const img = _mapP5.loadImage(
          line(
            items[k].Text || "",
            items[k].Color || "white",
            items[k].Values || [],
            backColor
          )
        );
        if (img) {
          imageMap.set(k, img);
        }
        break;
      }
    }
  }
  mapRedraw = true;
};

export const zoom = (zoomin: boolean) => {
  scale += zoomin ? 0.05 : -0.05;

  if (scale > 3.0) {
    scale = 3.0;
  } else if (scale < 0.05) {
    scale = 0.05;
  }
  mapRedraw = true;
};

const _setMapState = () => {
  mapState = 0;
  for (const id in nodes) {
    switch (nodes[id].State) {
      case "high":
        mapState = 2;
        return;
      case "low":
        mapState = 1;
        break;
    }
  }
};

export const setMapReadOnly = (ro: boolean) => {
  readOnly = ro || settingsLock;
};

let player: HTMLAudioElement = new Audio();

const _checkBeep = async () => {
  if (player && player.onplaying) {
    return;
  }
  if (mapState < 1) {
    return;
  }
  if (mapState == 2 && beepHigh) {
    player.src = beepHigh;
    player.play();
    return;
  }
  if (beepLow) {
    player.src = beepLow;
    player.play();
    return;
  }
};

export const resetMap = () => {
  imageMap.clear();
};

export const deleteMap = () => {
  if (_mapP5) {
    _mapP5.remove();
    _mapP5 = undefined;
  }
};

export const grid = (g: number, test: boolean) => {
  const list = [];
  const mx = Math.ceil(mapSizeX / g);
  const my = Math.ceil(mapSizeY / g);
  const m = new Array(mx);
  for (let x = 0; x < m.length; x++) {
    m[x] = new Array(my);
    for (let y = 0; y < m[x].length; y++) {
      m[x][y] = false;
    }
  }
  for (const id in nodes) {
    let x = Math.max(Math.min(Math.ceil((nodes[id].X * 1.0) / g), mx - 1), 0);
    let y = Math.max(Math.min(Math.ceil((nodes[id].Y * 1.0) / g), my - 1), 0);
    while (m[x][y]) {
      x++;
      if (x >= mx) {
        y++;
        x = 0;
        if (y >= my) {
          y = 0;
          break;
        }
      }
    }
    m[x][y] = true;
    nodes[id].X = x * g;
    nodes[id].Y = y * g;
    list.push({
      ID: id,
      X: nodes[id].X,
      Y: nodes[id].Y,
    });
  }
  if (!test && list.length > 0) {
    UpdateNodePos(list);
  }
  mapRedraw = true;
};

export const horizontal = (selected: any) => {
  const list = [];
  if (!selected || selected.length < 2) {
    return;
  }
  selected.sort((a: any, b: any) => {
    return nodes[a].X - nodes[b].X;
  });
  const id0 = selected[0];
  let dx = nodes[selected[1]].X - nodes[id0].X;
  if (dx < 40) {
    dx = 40;
  }
  let idLast = "";
  for (const id of selected) {
    if (id != id0) {
      nodes[id].Y = nodes[id0].Y;
      nodes[id].X = nodes[idLast].X + dx;
      if (nodes[id].X > mapSizeX - 80) {
        nodes[id].X = mapSizeX - 80;
      }
      list.push({
        ID: id,
        X: nodes[id].X,
        Y: nodes[id].Y,
      });
    }
    idLast = id;
  }
  if (list.length > 0) {
    UpdateNodePos(list);
  }
  mapRedraw = true;
};

export const vertical = (selected: any) => {
  const list = [];
  if (!selected || selected.length < 2) {
    return;
  }
  selected.sort((a: any, b: any) => {
    return nodes[a].Y - nodes[b].Y;
  });
  const id0 = selected[0];
  let dy = nodes[selected[1]].Y - nodes[id0].Y;
  if (dy < 60) {
    dy = 60;
  }
  let idLast = "";
  for (const id of selected) {
    if (id != id0) {
      nodes[id].X = nodes[id0].X;
      nodes[id].Y = nodes[idLast].Y + dy;
      if (nodes[id].Y > mapSizeY - 80) {
        nodes[id].Y = mapSizeY - 80;
      }
      list.push({
        ID: id,
        X: nodes[id].X,
        Y: nodes[id].Y,
      });
    }
    idLast = id;
  }
  if (list.length > 0) {
    UpdateNodePos(list);
  }
  mapRedraw = true;
};

export const circle = (selected: any) => {
  const list = [];
  if (!selected || selected.length < 2) {
    return;
  }
  selected.sort((a: any, b: any) => {
    return nodes[a].X - nodes[b].X;
  });
  const c = 80 * selected.length;
  const r = Math.min(Math.trunc(c / 3.14 / 2), mapSizeX / 2 - 80);
  const cx = nodes[selected[0]].X + r;
  let cy = nodes[selected[0]].Y;
  if (cy - r < 0) {
    cy = 40 + r;
  }
  for (let i = 0; i < selected.length; i++) {
    const id = selected[i];
    const d = 180 - i * (360 / selected.length);
    const a = (d * Math.PI) / 180;
    nodes[id].X = Math.max(Math.trunc(r * Math.cos(a) + cx), 0);
    nodes[id].Y = Math.max(Math.trunc(r * Math.sin(a) + cy), 0);
    list.push({
      ID: id,
      X: nodes[id].X,
      Y: nodes[id].Y,
    });
  }
  if (list.length > 0) {
    UpdateNodePos(list);
  }
  mapRedraw = true;
};

export const setShowAllItems = (s: boolean) => {
  showAllItems = s;
  mapRedraw = true;
};

const getLineColor = (state: any) => {
  if (state === "high" || state === "low" || state === "warn") {
    return getStateColor(state);
  }
  return 250;
};

const getLinePos = (id: string, polling: string) => {
  if (id.startsWith("NET:")) {
    const a = id.split(":");
    if (a.length !== 2) {
      return undefined;
    }
    const net = networks[a[1]];
    if (!net || !net.Ports) {
      return undefined;
    }
    let pi = -1;
    for (let i = 0; i < net.Ports.length; i++) {
      if (net.Ports[i].ID === polling) {
        pi = i;
        break;
      }
    }
    if (pi < 0) {
      return undefined;
    }
    return {
      X: net.X + net.Ports[pi].X * 45 + 10 + 20,
      Y: net.Y + net.Ports[pi].Y * 55 + fontSize + 20 + 10,
    };
  }
  if (!nodes[id]) {
    return undefined;
  }
  return {
    X: nodes[id].X,
    Y: nodes[id].Y + 6,
  };
};

const isDark = (): boolean => {
  const e = document.querySelector("html");
  if (!e) {
    return false;
  }
  return e.classList.contains("dark");
};

const condCheck = (c: number) => {
  return mapState >= c || showAllItems;
};

const mapMain = (p5: P5) => {
  let startMouseX = 0;
  let startMouseY = 0;
  let lastMouseX = 0;
  let lastMouseY = 0;
  let dragMode = 0; // 0 : None , 1: Select , 2 :Move
  let oldDark = isDark();
  let draggedNetwork = "";
  const draggedNodes: any = [];
  const draggedItems: any = [];
  let clickInCanvas = false;
  p5.setup = () => {
    const c = p5.createCanvas(mapSizeX, mapSizeY);
    c.mousePressed(canvasMousePressed);
  };
  p5.draw = () => {
    const dark = isDark();
    if (dark != oldDark) {
      mapRedraw = true;
      oldDark = dark;
    }
    if (!mapRedraw) {
      return;
    }
    if (scale != 1.0) {
      p5.scale(scale);
    }
    mapRedraw = false;
    p5.clear();
    p5.removeElements();
    p5.background(dark ? 23 : 252);
    if (_backImage) {
      if (backImage.Width) {
        p5.image(
          _backImage,
          backImage.X,
          backImage.Y,
          backImage.Width,
          backImage.Height
        );
      } else {
        p5.image(_backImage, backImage.X, backImage.Y);
      }
    }
    for (const k in networks) {
      p5.push();
      p5.translate(networks[k].X, networks[k].Y);
      if (selectedNetwork === networks[k].ID) {
        p5.stroke("#02c");
      } else if (networks[k].Error != "") {
        p5.stroke("#cc3300");
      } else {
        p5.stroke("#999");
      }
      p5.fill("rgba(23,23,23,0.9)");
      p5.rect(0, 0, networks[k].W, networks[k].H);
      p5.stroke("#999");
      p5.textFont("Roboto");
      p5.textSize(fontSize);
      p5.fill("#eee");
      p5.text(networks[k].Name, 5, fontSize + 5);
      if (!networks[k].Ports || networks[k].Ports.length < 1) {
        if (networks[k].Error !== "") {
          p5.fill("#cc3300");
          p5.text(networks[k].Error, 15, fontSize * 2 + 15);
        } else {
          p5.fill("#11ee00");
          p5.text("Check network node...", 15, fontSize * 2 + 15);
        }
      } else if (portImage) {
        p5.textSize(6);
        for (const p of networks[k].Ports) {
          const x = p.X * 45 + 10;
          const y = p.Y * 55 + fontSize + 15;
          p5.image(portImage, x, y, 40, 40);
          p5.fill(p.State === "up" ? "#11ee00" : " #999");
          p5.circle(x + 4, y + 4, 8);
          p5.fill("#eee");
          p5.text(p.Name, x, y + 40 + 10);
        }
      }
      p5.pop();
    }
    for (const k in lines) {
      const lp1 = getLinePos(lines[k].NodeID1, lines[k].PollingID1);
      if (!lp1) {
        continue;
      }
      const lp2 = getLinePos(lines[k].NodeID2, lines[k].PollingID2);
      if (!lp2) {
        continue;
      }
      const x1 = lp1.X;
      const x2 = lp2.X;
      const y1 = lp1.Y;
      const y2 = lp2.Y;
      const xm = (x1 + x2) / 2;
      const ym = (y1 + y2) / 2;
      p5.push();
      p5.strokeWeight(lines[k].Width || 1);
      p5.stroke(getStateColor(lines[k].State1));
      p5.line(x1, y1, xm, ym);
      p5.stroke(getStateColor(lines[k].State2));
      p5.line(xm, ym, x2, y2);
      if (lines[k].Info) {
        const color: any = getLineColor(lines[k].State);
        const dx = Math.abs(x1 - x2);
        const dy = Math.abs(y1 - y2);
        p5.textFont("Roboto");
        p5.textSize(fontSize);
        p5.fill(color);
        if (dx === 0 || dy / dx > 0.8) {
          p5.text(lines[k].Info, xm + 10, ym);
        } else {
          p5.text(lines[k].Info, xm - dx / 4, ym + 20);
        }
      }
      p5.pop();
    }
    for (const k in items) {
      if (items[k].Type < 4 && !condCheck(items[k].Cond)) {
        continue;
      }
      p5.push();
      p5.translate(items[k].X, items[k].Y);
      if (selectedDrawItems.includes(items[k].ID)) {
        if (dark) {
          p5.fill("rgba(23,23,23,0.9)");
          p5.stroke("#ccc");
        } else {
          p5.fill("rgba(252,252,252,0.9)");
          p5.stroke("#333");
        }
        const w = items[k].W + 10;
        const h = items[k].H + 10;
        p5.rect(-5, -5, w, h);
      }
      switch (items[k].Type) {
        case 0: // rect
          p5.fill(items[k].Color);
          p5.stroke("rgba(23,23,23,0.9)");
          p5.rect(0, 0, items[k].W, items[k].H);
          break;
        case 1: // ellipse
          p5.fill(items[k].Color);
          p5.stroke("rgba(23,23,23,0.9)");
          p5.ellipse(items[k].W / 2, items[k].H / 2, items[k].W, items[k].H);
          break;
        case 2: // text
        case 4: // Polling
          p5.textSize(items[k].Size || 12);
          p5.fill(items[k].Color);
          p5.text(
            items[k].Text,
            0,
            0,
            items[k].Size * items[k].Text.length + 10,
            items[k].Size + 10
          );
          break;
        case 3: // Image
          if (imageMap.has(items[k].ID)) {
            p5.image(imageMap.get(items[k].ID), 0, 0, items[k].W, items[k].H);
          } else {
            p5.fill("#aaa");
            p5.rect(0, 0, items[k].W, items[k].H);
          }
          break;
        case 5: {
          // Gauge
          const x = items[k].W / 2;
          const y = items[k].H / 2;
          const r0 = items[k].W;
          const r1 = items[k].W - items[k].Size;
          const r2 = items[k].W - items[k].Size * 4;
          p5.noStroke();
          p5.fill(dark ? "#eee" : "#333");
          p5.arc(x, y, r0, r0, 5 * p5.QUARTER_PI, -p5.QUARTER_PI);
          if (items[k].Value > 0) {
            p5.fill(items[k].Color);
            p5.arc(
              x,
              y,
              r0,
              r0,
              5 * p5.QUARTER_PI,
              -p5.QUARTER_PI -
                (p5.HALF_PI - p5.HALF_PI * Math.min(items[k].Value / 100, 1.0))
            );
          }
          p5.fill(dark ? 23 : 252);
          p5.arc(x, y, r1, r1, -p5.PI, 0);
          p5.textAlign(p5.CENTER);
          p5.textSize(8);
          p5.fill(dark ? "#eee" : "#333");
          p5.text(Number(items[k].Value).toFixed(3) + "%", x, y - 10);
          p5.textSize(items[k].Size);
          p5.text(items[k].Text || "", x, y + items[k].Size);
          p5.fill("#e31a1c");
          const angle = -p5.QUARTER_PI + (p5.HALF_PI * items[k].Value) / 100;
          const x1 = x + (r1 / 2) * p5.sin(angle);
          const y1 = y - (r1 / 2) * p5.cos(angle);
          const x2 = x + (r2 / 2) * p5.sin(angle) + 5 * p5.cos(angle);
          const y2 = y - (r2 / 2) * p5.cos(angle) + 5 * p5.sin(angle);
          const x3 = x + (r2 / 2) * p5.sin(angle) - 5 * p5.cos(angle);
          const y3 = y - (r2 / 2) * p5.cos(angle) - 5 * p5.sin(angle);
          p5.triangle(x1, y1, x2, y2, x3, y3);
        }
        case 6: // New Gauge,Line,Bar
        case 7:
        case 8:
          if (imageMap.has(k)) {
            p5.image(imageMap.get(k), 0, 0, items[k].W, items[k].H);
          }
          break;
      }
      p5.pop();
    }
    for (const k in nodes) {
      const icon = getIconCode(nodes[k].Icon);
      p5.push();
      p5.translate(nodes[k].X, nodes[k].Y);
      if (nodes[k].Image && imageMap.has(nodes[k].Image)) {
        const img = imageMap.get(nodes[k].Image);
        const w = 48 + 16;
        const h = img.height + 16 + fontSize;
        if (selectedNodes.includes(nodes[k].ID)) {
          if (dark) {
            p5.fill("rgba(23,23,23,0.9)");
          } else {
            p5.fill("rgba(252,252,252,0.9)");
          }
          p5.stroke(getStateColor(nodes[k].State));
          p5.rect(-w / 2, -h / 2, w, h);
        } else {
          const w = 40;
          if (dark) {
            p5.fill("rgba(23,23,23,0.9)");
            p5.stroke("rgba(23,23,23,0.9)");
          } else {
            p5.fill("rgba(252,252,252,0.9)");
            p5.stroke("rgba(252,252,252,0.9)");
          }
          p5.rect(-w / 2, -h / 2, w, h);
        }
        p5.tint(getStateColor(nodes[k].State));
        p5.image(img, -24, -h / 2 + 10, 48);
        p5.noTint();
        p5.textAlign(p5.CENTER, p5.CENTER);
        p5.textFont("Roboto");
        p5.textSize(fontSize);
        p5.textSize(fontSize);
        if (dark) {
          p5.fill(250);
        } else {
          p5.fill(23);
        }
        p5.text(nodes[k].Name, 0, img.height);
      } else {
        if (selectedNodes.includes(nodes[k].ID)) {
          if (dark) {
            p5.fill("rgba(23,23,23,0.9)");
          } else {
            p5.fill("rgba(252,252,252,0.9)");
          }
          p5.stroke(getStateColor(nodes[k].State));
          const w = iconSize + 16;
          p5.rect(-w / 2, -w / 2, w, w);
        } else {
          if (dark) {
            p5.fill("rgba(23,23,23,0.9)");
            p5.stroke("rgba(23,23,23,0.9)");
          } else {
            p5.fill("rgba(252,252,252,0.9)");
            p5.stroke("rgba(252,252,252,0.9)");
          }
          const w = iconSize - 8;
          p5.rect(-w / 2, -w / 2, w, w);
        }
        p5.textFont("Material Design Icons");
        p5.textSize(iconSize);
        p5.textAlign(p5.CENTER, p5.CENTER);
        p5.fill(getStateColor(nodes[k].State));
        p5.text(icon, 0, 0);
        p5.textFont("Roboto");
        p5.textSize(fontSize);
        if (dark) {
          p5.fill(250);
        } else {
          p5.fill(23);
        }
        p5.text(nodes[k].Name, 0, 32);
      }
      p5.pop();
    }
    if (dragMode === 1) {
      let x = startMouseX;
      let y = startMouseY;
      let w = lastMouseX - startMouseX;
      let h = lastMouseY - startMouseY;
      if (startMouseX > lastMouseX) {
        x = lastMouseX;
        w = startMouseX - lastMouseX;
      }
      if (startMouseY > lastMouseY) {
        y = lastMouseY;
        h = startMouseY - lastMouseY;
      }
      p5.push();
      p5.fill("rgba(250,250,250,0.6)");
      p5.stroke(0);
      p5.rect(x, y, w, h);
      p5.pop();
    }
  };

  p5.mouseDragged = (e: MouseEvent) => {
    if (readOnly && !clickInCanvas) {
      return true;
    }
    if (p5.mouseButton === p5.RIGHT ) {
      return true
    }
    if (dragMode === 0) {
      if (
        selectedNodes.length > 0 ||
        selectedDrawItems.length > 0 ||
        selectedNetwork !== ""
      ) {
        dragMode = 2;
      } else {
        dragMode = 1;
      }
    }
    if (dragMode === 1) {
      dragSelectNodes();
    } else if (dragMode === 2 && lastMouseX) {
      dragMoveNodes();
    }
    lastMouseX = p5.mouseX / scale;
    lastMouseY = p5.mouseY / scale;
    return true;
  };

  let selectedNetwork2 = "";
  const checkLine = () => {
    if (!p5.keyIsDown(p5.SHIFT)) {
      return false;
    }
    if (selectedNetwork != "") {
      if (setSelectNode(true)) {
        return true;
      }
      if (setSelectNetwork(true)) {
        return true;
      }
    } else if (selectedNodes.length === 1) {
      if (setSelectNode(true)) {
        return true;
      }
      if (setSelectNetwork(false)) {
        return true;
      }
    }
    return false;
  };

  const canvasMousePressed = () => {
    if (readOnly) {
      return true;
    }
    clickInCanvas = true;
    mapRedraw = true;
    if (checkLine()) {
      editLine();
      selectedNodes.length = 0;
      selectedNetwork = "";
      selectedNetwork2 = "";
      return false;
    } else if (p5.keyIsDown(p5.ALT)) {
      setSelectNode(true);
    } else if (dragMode !== 3) {
      setSelectNode(false);
      setSelectItem();
      setSelectNetwork(false);
    }
    lastMouseX = p5.mouseX / scale;
    lastMouseY = p5.mouseY / scale;
    startMouseX = p5.mouseX / scale;
    startMouseY = p5.mouseY / scale;
    dragMode = 0;
    return false;
  };

  p5.mouseReleased = (e) => {
    if (readOnly) {
      return true;
    }
    mapRedraw = true;
    if (!clickInCanvas) {
      selectedNodes.length = 0;
      selectedDrawItems.length = 0;
      return true;
    }
    if (
      p5.mouseButton === p5.RIGHT &&
      selectedNodes.length + selectedDrawItems.length < 2
    ) {
      if (mapCallBack) {
        mapCallBack({
          Cmd: "contextMenu",
          Node: selectedNodes[0] || "",
          DrawItem: selectedDrawItems[0] || "",
          Network: selectedNetwork || "",
          x: p5.winMouseX,
          y: p5.winMouseY,
        });
      }
    }
    if (p5.mouseButton === p5.RIGHT && selectedNodes.length > 1) {
      if (mapCallBack) {
        mapCallBack({
          Cmd: "formatNodes",
          Nodes: selectedNodes,
          x: p5.winMouseX,
          y: p5.winMouseY,
        });
      }
    }
    clickInCanvas = false;
    if (dragMode === 0 || dragMode === 3) {
      dragMode = 0;
      return false;
    }
    if (dragMode === 1) {
      if (selectedNodes.length > 0 || selectedDrawItems.length > 0) {
        dragMode = 3;
      } else {
        dragMode = 0;
      }
      return false;
    }
    if (draggedNodes.length > 0) {
      updateNodesPos();
    }
    if (draggedItems.length > 0) {
      updateItemsPos();
    }
    if (draggedNetwork !== "") {
      UpdateNetworkPos({
        ID: draggedNetwork,
        X: networks[draggedNetwork].X,
        Y: networks[draggedNetwork].Y,
      });
      draggedNetwork = "";
    }
    return false;
  };

  p5.keyReleased = () => {
    if (readOnly) {
      return true;
    }
    if (p5.keyCode === p5.DELETE || p5.keyCode === p5.BACKSPACE) {
      // Delete
      if (selectedNodes.length > 0) {
        deleteNodes();
      } else if (selectedDrawItems.length > 0) {
        deleteDrawItems();
      } else if (selectedNetwork != "") {
        deleteNetwork();
      }
      return true;
    }
    if (p5.keyCode === p5.ENTER) {
      p5.doubleClicked();
      return true;
    }
    switch (p5.key) {
      case "u":
      case "U":
        resizeDrawItem(1);
        break;
      case "d":
      case "D":
        resizeDrawItem(-1);
        break;
    }
    return true;
  };

  p5.doubleClicked = () => {
    if (selectedNodes.length === 1) {
      nodeDoubleClicked();
    } else if (selectedDrawItems.length === 1) {
      itemDoubleClicked();
    } else if (selectedNetwork !== "") {
      networkDoubleClicked();
    }
    return true;
  };

  const resizeDrawItem = (add: number) => {
    if (selectedDrawItems.length < 1) {
      return;
    }
    selectedDrawItems.forEach((id: any) => {
      if (items[id]) {
        switch (items[id].Type) {
          case 2:
          case 4:
            if (items[id].Size > 1) {
              items[id].Size += add;
            }
            items[id].W = items[id].Size * items[id].Text.length;
            items[id].H = items[id].Size;
            break;
          case 5:
            if (items[id].Size > 1) {
              items[id].Size += add;
            }
            items[id].H = items[id].Size * 10;
            items[id].W = items[id].Size * 10;
            break;
          case 6: // New Gauge
            if (items[id].H > 20) {
              items[id].H += add;
            }
            items[id].W = items[id].H;
            break;
          case 7: // Bar
          case 8: // Line
            if (items[id].H > 20) {
              items[id].H += add;
            }
            items[id].W = items[id].H * 4;
            break;
          default:
            items[id].W += add * 5;
            items[id].H += add * 5;
            items[id].Size += add;
            if (items[id].W < 10) {
              items[id].W = 10;
            }
            if (items[id].H < 10) {
              items[id].H = 10;
            }
            if (items[id].Size < 5) {
              items[id].Size = 5;
            }
        }
        UpdateDrawItem(items[id]);
      }
    });
    mapRedraw = true;
  };

  const checkNodePos = (n: any) => {
    if (n.X < 16) {
      n.X = 16;
    }
    if (n.Y < 16) {
      n.Y = 16;
    }
    if (n.X > mapSizeX) {
      n.X = mapSizeX - 16;
    }
    if (n.Y > mapSizeY) {
      n.Y = mapSizeY - 16;
    }
  };
  const checkItemPos = (i: any) => {
    if (i.X < 16) {
      i.X = 16;
    }
    if (i.Y < 16) {
      i.Y = 16;
    }
    if (i.X > mapSizeX - i.W) {
      i.X = mapSizeX - i.W;
    }
    if (i.Y > mapSizeY - i.H) {
      i.Y = mapSizeY - i.H;
    }
  };
  const dragMoveNodes = () => {
    selectedNodes.forEach((id: any) => {
      if (nodes[id]) {
        nodes[id].X += Math.trunc(p5.mouseX / scale - lastMouseX);
        nodes[id].Y += Math.trunc(p5.mouseY / scale - lastMouseY);
        checkNodePos(nodes[id]);
        if (!draggedNodes.includes(id)) {
          draggedNodes.push(id);
        }
      }
    });
    selectedDrawItems.forEach((id: any) => {
      if (items[id]) {
        items[id].X += Math.trunc(p5.mouseX / scale - lastMouseX);
        items[id].Y += Math.trunc(p5.mouseY / scale - lastMouseY);
        checkItemPos(items[id]);
        if (!draggedItems.includes(id)) {
          draggedItems.push(id);
        }
      }
    });
    if (selectedNetwork !== "" && networks[selectedNetwork]) {
      networks[selectedNetwork].X += Math.trunc(p5.mouseX / scale - lastMouseX);
      networks[selectedNetwork].Y += Math.trunc(p5.mouseY / scale - lastMouseY);
      draggedNetwork = selectedNetwork;
    }
    mapRedraw = true;
  };

  const dragSelectNodes = () => {
    selectedNodes.length = 0;
    const sx = startMouseX < lastMouseX ? startMouseX : lastMouseX;
    const sy = startMouseY < lastMouseY ? startMouseY : lastMouseY;
    const lx = startMouseX > lastMouseX ? startMouseX : lastMouseX;
    const ly = startMouseY > lastMouseY ? startMouseY : lastMouseY;
    for (const k in nodes) {
      if (
        nodes[k].X > sx &&
        nodes[k].X < lx &&
        nodes[k].Y > sy &&
        nodes[k].Y < ly
      ) {
        selectedNodes.push(nodes[k].ID);
      }
    }
    selectedDrawItems.length = 0;
    for (const k in items) {
      if (
        items[k].X > sx &&
        items[k].X < lx &&
        items[k].Y > sy &&
        items[k].Y < ly
      ) {
        selectedDrawItems.push(items[k].ID);
      }
    }
    mapRedraw = true;
  };

  const setSelectNode = (bMulti: boolean) => {
    const l = selectedNodes.length;
    const x = p5.mouseX / scale;
    const y = p5.mouseY / scale;
    for (const k in nodes) {
      if (
        nodes[k].X + 32 > x &&
        nodes[k].X - 32 < x &&
        nodes[k].Y + 32 > y &&
        nodes[k].Y - 32 < y
      ) {
        if (selectedNodes.includes(nodes[k].ID)) {
          if (bMulti) {
            const i = selectedNodes.indexOf(nodes[k].ID);
            selectedNodes.splice(i, 1);
          }
          return false;
        }
        if (!bMulti) {
          selectedNodes.length = 0;
        }
        selectedNodes.push(nodes[k].ID);
        return true;
      }
    }
    if (!bMulti) {
      selectedNodes.length = 0;
    }
    return l !== selectedNodes.length;
  };
  // 描画アイテムを選択する
  const setSelectItem = () => {
    const x = p5.mouseX / scale;
    const y = p5.mouseY / scale;
    for (const k in items) {
      const w = items[k].W + 10;
      const h = items[k].H + 10;
      if (
        items[k].X + w > x &&
        items[k].X - 10 < x &&
        items[k].Y + h > y &&
        items[k].Y - 10 < y &&
        (condCheck(items[k].Cond) || showAllItems)
      ) {
        if (selectedDrawItems.includes(items[k].ID)) {
          return;
        }
        selectedDrawItems.push(items[k].ID);
        return;
      }
    }
    selectedDrawItems.length = 0;
  };
  // Networkを選択する
  const setSelectNetwork = (second: boolean) => {
    const x = p5.mouseX / scale;
    const y = p5.mouseY / scale;
    for (const k in networks) {
      const w = networks[k].W + 10;
      const h = networks[k].H;
      if (
        networks[k].X + w > x &&
        networks[k].X - 10 < x &&
        networks[k].Y + h > y &&
        networks[k].Y - 10 < y
      ) {
        if (second) {
          selectedNetwork2 = networks[k].ID;
        } else {
          selectedNetwork = networks[k].ID;
        }
        return true;
      }
    }
    selectedNetwork = "";
    return false;
  };
  // ノードを削除する
  const deleteNodes = () => {
    if (mapCallBack) {
      mapCallBack({
        Cmd: "deleteNodes",
        Param: selectedNodes,
      });
      selectedNodes.length = 0;
    }
  };
  // 描画アイテムを削除する
  const deleteDrawItems = () => {
    if (mapCallBack) {
      mapCallBack({
        Cmd: "deleteDrawItems",
        Param: selectedDrawItems,
      });
      selectedDrawItems.length = 0;
    }
  };
  // ネットワークを削除する
  const deleteNetwork = () => {
    if (mapCallBack) {
      mapCallBack({
        Cmd: "deleteNetwork",
        Param: selectedNetwork,
      });
      selectedNetwork = "";
    }
  };
  // Nodeの位置を保存する
  const updateNodesPos = () => {
    const list: any = [];
    draggedNodes.forEach((id: any) => {
      if (nodes[id]) {
        // 位置を保存するノード
        list.push({
          ID: id,
          X: Math.trunc(nodes[id].X),
          Y: Math.trunc(nodes[id].Y),
        });
      }
    });
    UpdateNodePos(list);
    draggedNodes.length = 0;
  };
  // 描画アイテムの位置を保存する
  const updateItemsPos = () => {
    const list: any = [];
    draggedItems.forEach((id: any) => {
      if (items[id]) {
        // 位置を保存するノード
        list.push({
          ID: id,
          X: Math.trunc(items[id].X),
          Y: Math.trunc(items[id].Y),
        });
      }
    });
    UpdateDrawItemPos(list);
    draggedItems.length = 0;
  };
  // nodeをダブルクリックした場合
  const nodeDoubleClicked = () => {
    if (mapCallBack) {
      mapCallBack({
        Cmd: "nodeDoubleClicked",
        Param: selectedNodes[0],
      });
    }
  };
  // itemをダブルクリックした場合
  const itemDoubleClicked = () => {
    if (mapCallBack) {
      mapCallBack({
        Cmd: "itemDoubleClicked",
        Param: selectedDrawItems[0],
      });
    }
  };
  // networkをダブルクリックした場合
  const networkDoubleClicked = () => {
    if (mapCallBack) {
      mapCallBack({
        Cmd: "networkDoubleClicked",
        Param: selectedNetwork,
      });
    }
  };
  // lineの編集
  const editLine = () => {
    if (selectedNetwork != "") {
      selectedNodes.push("NET:" + selectedNetwork);
    }
    if (selectedNetwork2 != "") {
      selectedNodes.push("NET:" + selectedNetwork2);
    }
    if (selectedNodes.length !== 2) {
      return;
    }
    if (mapCallBack) {
      mapCallBack({
        Cmd: "editLine",
        Param: selectedNodes,
      });
    }
    selectedNodes.length = 0;
    mapRedraw = true;
  };
};
