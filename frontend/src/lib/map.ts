import P5 from 'p5';
import {getIconCode,getStateColor} from  './common'
import {
  GetSettings,
  GetNodes,
  GetLines,
  GetDrawItems,
  GetBackImage,
  UpdateDrawItemPos,
  UpdateNodePos,
  GetImage,
  GetMapConf,
  GetNotifyConf,
} from "../../wailsjs/go/main/App"
import type { datastore } from 'wailsjs/go/models';
import {gauge,line,bar} from './chart/drawitem';

const MAP_SIZE_X = window.screen.width > 4000 ? 5000 : 2500;
const MAP_SIZE_Y = 5000;
let mapRedraw = true;
let readOnly = false;

let mapCallBack :any = undefined;

let nodes :any = {};
let lines :any = [];
let items :any = {}
let backImage: datastore.BackImageEnt = {
  X:0,
  Y:0,
  Width:0,
  Height: 0,
  Path: '',
};

let _backImage:any = undefined; 

let fontSize = 12;
let iconSize = 32;

const selectedNodes :any = [];
const selectedDrawItems :any = [];

const imageMap = new Map();
let mapState = 0;
let showAllItems = false;

let _mapP5 :P5 | undefined  = undefined;
let beepHigh :any= undefined;
let beepLow :any = undefined;
let scale = 1.0;

export const initMAP = async (div:HTMLElement,cb :any) => {
  const settings = await GetSettings();
  const notifyConf = await GetNotifyConf();
  beepHigh = notifyConf.BeepHigh;
  beepLow = notifyConf.BeepLow;
  
  mapCallBack =cb;
  readOnly = settings.Lock != "";
  mapRedraw = false;
  if (_mapP5 != undefined) {
    return
  }
  div.oncontextmenu = (e) => {
    e.preventDefault()
  }
  _mapP5 = new P5(mapMain, div)
}


let lastBackImagePath = "";

export const updateMAP = async () => {
  const dark = isDark();
  const mapConf = await GetMapConf();
  const z = mapConf.IconSize ||  3;
  iconSize = 8 + z * 8;
  fontSize = 6 + z * 2; 
  nodes = await GetNodes();
  lines = await GetLines();
  items = await GetDrawItems() || {};
  backImage = await GetBackImage();
  if (_mapP5 != undefined){
    if (backImage.Path != lastBackImagePath) {
      if( backImage.Path) {
        _mapP5.loadImage(await GetImage(backImage.Path),(img)=>{
          _backImage = img;
          mapRedraw = true;
        },()=>{});
      } else {
        _backImage = null;
        mapRedraw = true;
      }
      lastBackImagePath = backImage.Path;
    }
  }
  _setMapState();
  _checkBeep();
  const backColor = _mapP5 ? dark ? _mapP5.color(23).toString() : _mapP5.color(252).toString() : "#333"; 
  for(const k in items) {
    switch (items[k].Type) {
    case 3:
      if (!imageMap.has(items[k].ID) && _mapP5 != undefined) {
        _mapP5.loadImage(await GetImage(items[k].Path),(img)=>{
          imageMap.set(items[k].ID,img);
          mapRedraw = true;
        },(e)=>{});
      }  
      break;
    case 2:
    case 4:
      items[k].W = items[k].Size *  items[k].Text.length;
      items[k].H = items[k].Size;
      if(!dark) {
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
      case 6: // New Gauge
        items[k].W = items[k].H
        if( _mapP5) {
          _mapP5.loadImage(
            gauge(
              items[k].Text || '',
              items[k].Value || 0,
              backColor
            ),
            (img) => {
              imageMap.set(k,img)
              mapRedraw = true;
            }
          )}
        break
      case 7: // Bar
        items[k].W = items[k].H * 4
        if (_mapP5) {
          _mapP5.loadImage(
            bar(
              items[k].Text || '',
              items[k].Color || 'white',
              items[k].Value || 0,
              backColor
            ),
            (img) => {
              imageMap.set(k,img)
              mapRedraw = true;
            }
          )
        }
        break
      case 8: // Line
        items[k].W = items[k].H * 4
        if(_mapP5) {
          _mapP5.loadImage(
            line(
              items[k].Text || '',
              items[k].Color || 'white',
              items[k].Values || [],
              backColor
            ),
            (img) => {
              imageMap.set(k,img)
              mapRedraw = true;
            }
          )
        }
        break
    } 
  }
  mapRedraw = true;
}

export const zoom = (zoomin:boolean) => {
  scale += zoomin ? 0.05 : -0.05;

  if(scale>3.0) {
    scale = 3.0;
  } else if (scale < 0.05) {
    scale = 0.05;
  }
  mapRedraw = true;
} 

const _setMapState = () => {
  mapState= 0;
  for (const id in nodes) {
    switch(nodes[id].State) {
      case "high":
        mapState = 2;
        return;
      case "low":
        mapState = 1;
        break;
    }
  }
}

let player: HTMLAudioElement = new Audio();

const _checkBeep = async () => {
  if (player && player.onplaying ) {
    return;
  }
  if (mapState < 1) {
    return;
  }
  if(mapState == 2 && beepHigh) {
    player.src = beepHigh;
    player.play();
    return;
  }
  if (beepLow) {
    player.src = beepLow;
    player.play();
    return;
  }
}

export const resetMap = () => {
  imageMap.clear();
}

export const deleteMap = () => {
  if(_mapP5) {
    _mapP5.remove();
    _mapP5 = undefined;
  }
}

export const grid = (g:number,test:boolean) => {
  const list = [];
  const mx = Math.ceil(MAP_SIZE_X/g); 
  const my = Math.ceil(MAP_SIZE_Y/g);
  const m = new Array(mx);
  for(let x = 0; x<m.length;x++) {
    m[x] = new Array(my);
    for(let y=0;y < m[x].length;y++) {
      m[x][y] = false;
    }
  }
  for (const id in nodes) {
    let x =  Math.max(Math.min(Math.ceil((nodes[id].X * 1.0) / g),mx-1),0);
    let y =  Math.max(Math.min(Math.ceil((nodes[id].Y * 1.0) / g),my-1),0);
    while(m[x][y]) {
      x++;
      if (x >= mx) {
        y++;
        x = 0;
        if(y >= my) {
          y = 0;
          break;
        }
      }
    }
    m[x][y] = true;
    nodes[id].X = x *  g;
    nodes[id].Y = y *  g;
    list.push({
      ID: id,
      X: nodes[id].X,
      Y: nodes[id].Y,
    })
  }
  if (!test && list.length > 0) {
    UpdateNodePos(list);
  }
  mapRedraw = true;
};

export const setShowAllItems = (s:boolean) => {
  showAllItems = s;
  mapRedraw = true;
}


const getLineColor = (state:any) => {
  if (state === 'high' || state === 'low' || state === 'warn') {
    return getStateColor(state)
  }
  return 250
}

const isDark = () :boolean => {
  const  e = document.querySelector("html");
  if (!e) {
    return false;
  }
  return e.classList.contains("dark");
}

const condCheck = (c:number) => {
  return mapState >= c || showAllItems;
}

const mapMain = (p5:P5) => {
  let startMouseX = 0;
  let startMouseY = 0;
  let lastMouseX = 0;
  let lastMouseY = 0;
  let dragMode  = 0; // 0 : None , 1: Select , 2 :Move
  let oldDark = false;
  const draggedNodes :any = [];
  const draggedItems :any = [];
  let clickInCanvas = false;
  p5.setup = () => {
    const c = p5.createCanvas(MAP_SIZE_X, MAP_SIZE_Y);
    c.mousePressed(canvasMousePressed);
  }

  p5.draw = () => {
    const dark = isDark();
    if (dark != oldDark) {
      mapRedraw = true;
      oldDark = dark;
    }
    if (!mapRedraw){
      return;
    }
    if(scale != 1.0) {
      p5.scale(scale);
    }
    mapRedraw = false;
    p5.background(dark ? 23 : 252 );
    if(_backImage){
      if(backImage.Width){
        p5.image(_backImage,backImage.X,backImage.Y,backImage.Width,backImage.Height);
      }else {
        p5.image(_backImage,backImage.X,backImage.Y);
      }
    }
    for (const k in lines) {
      if (!nodes[lines[k].NodeID1] || !nodes[lines[k].NodeID2]) {
        continue;
      }
      const x1 = nodes[lines[k].NodeID1].X;
      const x2 = nodes[lines[k].NodeID2].X;
      const y1 = nodes[lines[k].NodeID1].Y;
      const y2 = nodes[lines[k].NodeID2].Y;
      const xm = (x1 + x2) / 2;
      const ym = (y1 + y2) / 2;
      p5.push();
      p5.strokeWeight(lines[k].Width || 1 );
      p5.stroke(getStateColor(lines[k].State1));
      p5.line(x1, y1, xm, ym)
      p5.stroke(getStateColor(lines[k].State2));
      p5.line(xm, ym, x2, y2);
      if (lines[k].Info) {
        const color :any = getLineColor(lines[k].State);
        const dx = Math.abs(x1-x2);
        const dy = Math.abs(y1-y2);
        p5.textFont('Roboto');
        p5.textSize(fontSize);
        p5.fill(color);
        if (dx === 0 || dy/dx > 0.8) {
          p5.text(lines[k].Info, xm + 10, ym);
        } else {
          p5.text(lines[k].Info, xm - dx/4, ym + 20);
        }
      }
      p5.pop();
    }
    for (const k in items) {
      if(items[k].Type < 4 && !condCheck(items[k].Cond)) {
        continue;
      }
      p5.push();
      p5.translate(items[k].X, items[k].Y);
      if (selectedDrawItems.includes(items[k].ID) ) {
        if(dark) {
          p5.fill('rgba(23,23,23,0.9)');
          p5.stroke('#ccc');
        } else {
          p5.fill('rgba(252,252,252,0.9)');
          p5.stroke('#333');
        }
        const w =  items[k].W +10;
        const h =  items[k].H +10;
        p5.rect(-5, -5, w, h);
      }
      switch (items[k].Type) {
      case 0: // rect
        p5.fill(items[k].Color);
        p5.stroke('rgba(23,23,23,0.9)');
        p5.rect(0,0,items[k].W, items[k].H);
        break;
      case 1: // ellipse
        p5.fill(items[k].Color);
        p5.stroke('rgba(23,23,23,0.9)');
        p5.ellipse(items[k].W/2,items[k].H/2,items[k].W, items[k].H);
        break
      case 2: // text
      case 4: // Polling
        p5.textSize(items[k].Size || 12)
        p5.fill(items[k].Color)
        p5.text(items[k].Text, 0, 0,items[k].Size *  items[k].Text.length + 10, items[k].Size + 10)
        break
      case 3: // Image
        if (imageMap.has(items[k].ID)) {
          p5.image(imageMap.get(items[k].ID),0,0,items[k].W,items[k].H);
        } else {
          p5.fill("#aaa");
          p5.rect(0,0,items[k].W, items[k].H);
        }
        break
      case 5: { // Gauge
          const x = items[k].W / 2;
          const y = items[k].H / 2;
          const r0 = items[k].W;
          const r1 = (items[k].W - items[k].Size);
          const r2 = (items[k].W - items[k].Size *4)
          p5.noStroke();
          p5.fill(dark ? '#eee' : '#333');
          p5.arc(x, y, r0, r0, 5*p5.QUARTER_PI, -p5.QUARTER_PI);
          if(items[k].Value > 0){
            p5.fill(items[k].Color);
            p5.arc(x, y, r0, r0, 5*p5.QUARTER_PI, -p5.QUARTER_PI - (p5.HALF_PI - p5.HALF_PI * Math.min(items[k].Value/100,1.0)));
          }
          p5.fill(dark ? 23 :252);
          p5.arc(x, y, r1, r1, -p5.PI, 0);
          p5.textAlign(p5.CENTER);
          p5.textSize(8);
          p5.fill(dark ? '#eee' :'#333');
          p5.text( Number(items[k].Value).toFixed(3) + '%', x, y - 10);
          p5.textSize(items[k].Size);
          p5.text( items[k].Text || "", x, y + items[k].Size);
          p5.fill('#e31a1c');
          const angle = -p5.QUARTER_PI + (p5.HALF_PI * items[k].Value/100);
          const x1 = x + r1/2 * p5.sin(angle);
          const y1 = y - r1/2 * p5.cos(angle);
          const x2 = x + r2/2 * p5.sin(angle) + 5  * p5.cos(angle);
          const y2 = y - r2/2 * p5.cos(angle) + 5  * p5.sin(angle);
          const x3 = x + r2/2 * p5.sin(angle) - 5  * p5.cos(angle);
          const y3 = y - r2/2 * p5.cos(angle) - 5  * p5.sin(angle);
          p5.triangle(x1, y1, x2, y2, x3, y3);
        }
        case 6: // New Gauge,Line,Bar
        case 7:
        case 8:
          if (imageMap.has(k)) {
            p5.image(imageMap.get(k), 0, 0, items[k].W, items[k].H)
          }
          break
      }
      p5.pop();
    }
    for (const k in nodes) {
      const icon = getIconCode(nodes[k].Icon);
      p5.push()
      p5.translate(nodes[k].X, nodes[k].Y)
      if (selectedNodes.includes(nodes[k].ID)) {
        if(dark) {
          p5.fill('rgba(23,23,23,0.9)')
        } else {
          p5.fill('rgba(252,252,252,0.9)')
        }
        p5.stroke(getStateColor(nodes[k].State))
        const w = iconSize + 16
        p5.rect(-w/2, -w/2, w, w)
      } else {
        if (dark) {
          p5.fill('rgba(23,23,23,0.9)')
          p5.stroke('rgba(23,23,23,0.9)')
        } else {
          p5.fill('rgba(252,252,252,0.9)')
          p5.stroke('rgba(252,252,252,0.9)')
        }
        const w = iconSize - 8
        p5.rect(-w/2, -w/2, w, w)
      }
      p5.textFont('Material Design Icons')
      p5.textSize(iconSize)
      p5.textAlign(p5.CENTER, p5.CENTER)
      p5.fill(getStateColor(nodes[k].State))
      p5.text(icon, 0, 0)
      p5.textFont('Roboto')
      p5.textSize(fontSize)
      if(dark) {
        p5.fill(250)
      } else {
        p5.fill(23)
      }
      p5.text(nodes[k].Name, 0, 32)
      p5.pop()
    }
    if (dragMode === 1) {
      let x = startMouseX
      let y = startMouseY
      let w = lastMouseX  - startMouseX
      let h = lastMouseY  - startMouseY
      if (startMouseX > lastMouseX){
        x = lastMouseX
        w = startMouseX - lastMouseX
      }
      if (startMouseY > lastMouseY){
        y = lastMouseY
        h = startMouseY - lastMouseY
      }
      p5.push()
      p5.fill('rgba(250,250,250,0.6)')
      p5.stroke(0)
      p5.rect(x,y,w,h)
      p5.pop();
    } 
  }

  p5.mouseDragged = (e:MouseEvent) => {
    if (readOnly && !clickInCanvas) {
      return true
    }
    if (dragMode === 0) {
      if (selectedNodes.length > 0 || selectedDrawItems.length > 0 ){
        dragMode = 2
      } else {
        dragMode = 1
      }
    }
    if (dragMode === 1) {
      dragSelectNodes()
    } else if (dragMode === 2 && lastMouseX) {
      dragMoveNodes()
    }
    lastMouseX = p5.mouseX
    lastMouseY = p5.mouseY
    return true
  }

  const canvasMousePressed = () => {
    if (readOnly) {
      return true
    }
    clickInCanvas = true
    mapRedraw = true
    if (
      p5.keyIsDown(p5.SHIFT) &&
      selectedNodes.length === 1 &&
      setSelectNode(true)
    ) {
      editLine()
      selectedNodes.length = 0
      return false
    } else  if (dragMode !== 3) {
        setSelectNode(false)
        setSelectItem()
    }
    lastMouseX = p5.mouseX
    lastMouseY = p5.mouseY
    startMouseX = p5.mouseX
    startMouseY = p5.mouseY
    dragMode = 0
    return false
  }

  p5.mouseReleased = (e) => {
    if (readOnly) {
      return true
    }
    mapRedraw = true
    if(!clickInCanvas){
      selectedNodes.length = 0
      selectedDrawItems.length = 0
      return true
    }
    if(p5.mouseButton === p5.RIGHT && (selectedNodes.length + selectedDrawItems.length) < 2 ) {
      if (mapCallBack) {
        mapCallBack({
          Cmd: 'contextMenu',
          Node: selectedNodes[0] || '',
          DrawItem: selectedDrawItems[0] || '',
          x: p5.winMouseX,
          y: p5.winMouseY,
        })
      }
    }
    clickInCanvas = false 
    if (dragMode === 0 || dragMode === 3) {
      dragMode = 0
      return false
    }
    if (dragMode === 1) {
      if (selectedNodes.length > 0 || selectedDrawItems.length > 0 ){
        dragMode = 3
      } else {
        dragMode = 0
      }
      return false
    }
    if (draggedNodes.length > 0) {
      updateNodesPos()
    }
    if (draggedItems.length > 0) {
      updateItemsPos()
    }
    return false
  }

  p5.keyReleased = () => {
    if (readOnly) {
      return true
    }
    if (p5.keyCode === p5.DELETE || p5.keyCode === p5.BACKSPACE) {
      // Delete
      if (selectedNodes.length > 0){
        deleteNodes();
      } else if(selectedDrawItems.length > 0 ) {
        deleteDrawItems();
      }
    }
    if (p5.keyCode === p5.ENTER) {
      p5.doubleClicked()
    }
    return true
  }

  p5.doubleClicked = () => {
    if (selectedNodes.length === 1 ){
      nodeDoubleClicked()
    } else if (selectedDrawItems.length === 1 ){
      itemDoubleClicked()
    }
    return true
  }
  const checkNodePos = (n:any) => {
    if (n.X < 16) {
      n.X = 16
    }
    if (n.Y < 16) {
      n.Y = 16
    }
    if (n.X > MAP_SIZE_X) {
      n.X = MAP_SIZE_X - 16
    }
    if (n.Y > MAP_SIZE_Y) {
      n.Y = MAP_SIZE_Y - 16
    }
  }
  const checkItemPos = (i:any) => {
    if (i.X < 16) {
      i.X = 16
    }
    if (i.Y < 16) {
      i.Y = 16
    }
    if (i.X > MAP_SIZE_X - i.W) {
      i.X = MAP_SIZE_X - i.W
    }
    if (i.Y > MAP_SIZE_Y - i.H) {
      i.Y = MAP_SIZE_Y - i.H
    }
  }
  const dragMoveNodes = () => {
    selectedNodes.forEach((id:any) => {
      if (nodes[id]) {
        nodes[id].X += p5.mouseX - lastMouseX
        nodes[id].Y += p5.mouseY - lastMouseY
        checkNodePos(nodes[id])
        if (!draggedNodes.includes(id)) {
          draggedNodes.push(id)
        }
      }
    })
    selectedDrawItems.forEach((id:any) => {
      if (items[id]) {
        items[id].X += p5.mouseX - lastMouseX
        items[id].Y += p5.mouseY - lastMouseY
        checkItemPos(items[id])
        if (!draggedItems.includes(id)) {
          draggedItems.push(id)
        }
      }
    })
    mapRedraw = true
  }

  const dragSelectNodes = () => {
    selectedNodes.length = 0
    const sx = startMouseX < lastMouseX ? startMouseX : lastMouseX
    const sy = startMouseY < lastMouseY ? startMouseY : lastMouseY
    const lx = startMouseX > lastMouseX ? startMouseX : lastMouseX
    const ly = startMouseY > lastMouseY ? startMouseY : lastMouseY
    for (const k in nodes) {
      if (
        nodes[k].X > sx &&
        nodes[k].X < lx &&
        nodes[k].Y > sy &&
        nodes[k].Y < ly
      ) {
        selectedNodes.push(nodes[k].ID)
      }
    }
    selectedDrawItems.length = 0
    for (const k in items) {
      if (
        items[k].X > sx &&
        items[k].X < lx &&
        items[k].Y > sy &&
        items[k].Y < ly
      ) {
        selectedDrawItems.push(items[k].ID)
      }
    }
    mapRedraw = true
  }

  const setSelectNode = (bMulti:boolean) => {
    const l = selectedNodes.length
    for (const k in nodes) {
      if (
        nodes[k].X + 32 > p5.mouseX &&
        nodes[k].X - 32 < p5.mouseX &&
        nodes[k].Y + 32 > p5.mouseY &&
        nodes[k].Y - 32 < p5.mouseY
      ) {
        if (selectedNodes.includes(nodes[k].ID)) {
          return false
        }
        if (!bMulti) {
          selectedNodes.length = 0
        }
        selectedNodes.push(nodes[k].ID)
        return true
      }
    }
    if (!bMulti) {
      selectedNodes.length = 0
    }
    return l !== selectedNodes.length
  }
  // 描画アイテムを選択する
  const setSelectItem = () => {
    for (const k in items) {
      const w =  items[k].W +10
      const h =  items[k].H +10
      if (
        items[k].X + w > p5.mouseX &&
        items[k].X - 10 < p5.mouseX &&
        items[k].Y + h > p5.mouseY &&
        items[k].Y - 10 < p5.mouseY &&
        (condCheck(items[k].Cond) || showAllItems)
      ) {
        if (selectedDrawItems.includes(items[k].ID)) {
          return
        }
        selectedDrawItems.push(items[k].ID)
        return
      }
    }
    selectedDrawItems.length = 0
  }
  // ノードを削除する
  const deleteNodes = () => {
    if (mapCallBack){
      mapCallBack({
        Cmd: 'deleteNodes',
        Param: selectedNodes,
      })
      selectedNodes.length = 0
    }
  }
  // 描画アイテムを削除する
  const deleteDrawItems = () => {
    if (mapCallBack){
      mapCallBack({
        Cmd: 'deleteDrawItems',
        Param: selectedDrawItems,
      })
      selectedDrawItems.length = 0
    }
  }
  // Nodeの位置を保存する
  const updateNodesPos = () => {
    const list :any   = []
    draggedNodes.forEach((id:any) => {
      if (nodes[id]) {
        // 位置を保存するノード
        list.push({
          ID: id,
          X: nodes[id].X,
          Y: nodes[id].Y,
        })
      }
    })
    UpdateNodePos(list);
    draggedNodes.length = 0
  }
  // 描画アイテムの位置を保存する
  const updateItemsPos = () => {
    const list :any = []
    draggedItems.forEach((id:any) => {
      if (items[id]) {
        // 位置を保存するノード
        list.push({
          ID: id,
          X: items[id].X,
          Y: items[id].Y,
        })
      }
    })
    UpdateDrawItemPos(list);
    draggedItems.length = 0
  }
  // nodeをダブルクリックした場合
  const nodeDoubleClicked = () => {
    if (mapCallBack) {
      mapCallBack({
        Cmd: 'nodeDoubleClicked',
        Param: selectedNodes[0],
      })
    }
  }
  // itemをダブルクリックした場合
  const itemDoubleClicked = () => {
    if (mapCallBack) {
      mapCallBack({
        Cmd: 'itemDoubleClicked',
        Param: selectedDrawItems[0],
      })
    }
  }
  // lineの編集
  const editLine = () => {
    if (selectedNodes.length !== 2 ){
      return
    }
    if (mapCallBack) {
      mapCallBack({
        Cmd: 'editLine',
        Param: selectedNodes,
      })
    }
    selectedNodes.length = 0
    mapRedraw = true
  }
}
