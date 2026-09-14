import * as echarts from 'echarts'

// --- twFathom-style color palette ---
const COLOR_PRIMARY = '#00d2ff'
const COLOR_SUCCESS = '#10b981'
const COLOR_WARNING = '#f59e0b'
const COLOR_DANGER = '#ef4444'

/**
 * Modern Ring Gauge (twFathom Style)
 */
export const gauge = (title: string, val: number, backgroundColor: string): string => {
  const chart = echarts.init(null, null, {
    renderer: 'svg',
    ssr: true,
    width: 600,
    height: 600,
  })

  // Determine accent color by threshold
  const accentColor = val >= 90 ? COLOR_DANGER : val >= 80 ? COLOR_WARNING : COLOR_PRIMARY

  const option = {
    backgroundColor: 'transparent',
    series: [
      {
        type: 'gauge',
        startAngle: 210,
        endAngle: -30,
        min: 0,
        max: 100,
        splitNumber: 5,
        radius: '88%',
        progress: {
          show: true,
          roundCap: true,
          width: 18,
          itemStyle: {
            color: accentColor,
            shadowColor: accentColor + '88',
            shadowBlur: 12,
          },
        },
        pointer: {
          show: false,
        },
        axisLine: {
          roundCap: true,
          lineStyle: {
            width: 18,
            color: [[1, 'rgba(255, 255, 255, 0.08)']],
          },
        },
        axisTick: {
          distance: -28,
          length: 6,
          lineStyle: {
            color: 'rgba(255, 255, 255, 0.2)',
            width: 1.5,
          },
        },
        splitLine: {
          distance: -30,
          length: 12,
          lineStyle: {
            color: 'rgba(255, 255, 255, 0.35)',
            width: 2,
          },
        },
        axisLabel: {
          distance: -20,
          color: 'rgba(255, 255, 255, 0.5)',
          fontSize: 13,
          fontFamily: 'Inter, sans-serif',
        },
        title: {
          show: true,
          offsetCenter: [0, '62%'],
          fontSize: 22,
          fontWeight: 600,
          color: '#9ca3af',
          fontFamily: 'Outfit, Inter, sans-serif',
        },
        detail: {
          valueAnimation: true,
          fontSize: 48,
          fontWeight: 800,
          offsetCenter: [0, '5%'],
          formatter: '{value}%',
          color: '#f3f4f6',
          fontFamily: 'Outfit, Inter, sans-serif',
        },
        data: [
          {
            value: Number(val.toFixed(1)),
            name: title,
          },
        ],
      },
    ],
  }
  chart.setOption(option)
  return chart.getDataURL({ backgroundColor })
}

/**
 * Modern Area Sparkline (twFathom Style)
 */
export const line = (title: string, color: string, values: number[], backgroundColor: string): string => {
  const chart = echarts.init(null, null, {
    renderer: 'svg',
    ssr: true,
    width: 440,
    height: 120,
  })

  const strokeColor = color && color !== 'white' ? color : COLOR_PRIMARY
  const safeValues = values && values.length > 0 ? values : [0]
  const lastVal = safeValues[safeValues.length - 1]

  const option = {
    backgroundColor: 'transparent',
    title: {
      show: true,
      text: title,
      subtext: `${Number(lastVal).toFixed(1)}`,
      left: 14,
      top: 8,
      textStyle: {
        fontSize: 12,
        fontWeight: 700,
        color: '#9ca3af',
        fontFamily: 'Outfit, Inter, sans-serif',
      },
      subtextStyle: {
        fontSize: 15,
        fontWeight: 800,
        color: '#f3f4f6',
        fontFamily: 'Outfit, Inter, sans-serif',
      },
      itemGap: 2,
    },
    grid: {
      top: 42,
      left: 6,
      bottom: 6,
      right: 6,
    },
    xAxis: {
      show: false,
      type: 'category',
      data: safeValues.map((_, i) => i),
      boundaryGap: false,
    },
    yAxis: {
      show: false,
      type: 'value',
      min: 'dataMin',
    },
    series: [
      {
        type: 'line',
        smooth: 0.35,
        showSymbol: false,
        symbolSize: 6,
        lineStyle: {
          width: 2.5,
          color: strokeColor,
          shadowColor: strokeColor + '66',
          shadowBlur: 8,
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: strokeColor + '55' },
            { offset: 1, color: strokeColor + '00' },
          ]),
        },
        data: safeValues,
        markPoint: {
          symbol: 'circle',
          symbolSize: 7,
          itemStyle: {
            color: strokeColor,
            borderColor: '#ffffff',
            borderWidth: 1.5,
            shadowColor: strokeColor,
            shadowBlur: 8,
          },
          data: [{ coord: [safeValues.length - 1, lastVal] }],
        },
      },
    ],
  }
  chart.setOption(option)
  return chart.getDataURL({ backgroundColor })
}

/**
 * Modern Capsule Progress Bar (twFathom Style)
 */
export const bar = (title: string, color: string, value: number, backgroundColor: string): string => {
  const chart = echarts.init(null, null, {
    renderer: 'svg',
    ssr: true,
    width: 440,
    height: 90,
  })

  const barColor = color && color !== 'white' ? color : COLOR_PRIMARY
  const safeVal = Math.min(100, Math.max(0, value))

  const option = {
    backgroundColor: 'transparent',
    grid: {
      top: 36,
      left: 12,
      bottom: 12,
      right: 12,
    },
    title: {
      show: true,
      text: title,
      subtext: `${safeVal.toFixed(1)}%`,
      left: 12,
      top: 6,
      textStyle: {
        fontSize: 12,
        fontWeight: 700,
        color: '#9ca3af',
        fontFamily: 'Outfit, Inter, sans-serif',
      },
      subtextStyle: {
        fontSize: 14,
        fontWeight: 800,
        color: '#f3f4f6',
        fontFamily: 'Outfit, Inter, sans-serif',
      },
      itemGap: 2,
    },
    yAxis: {
      type: 'category',
      data: ['val'],
      show: false,
    },
    xAxis: {
      type: 'value',
      show: false,
      max: 100,
      min: 0,
    },
    series: [
      {
        data: [safeVal],
        type: 'bar',
        barWidth: 12,
        showBackground: true,
        backgroundStyle: {
          color: 'rgba(255, 255, 255, 0.08)',
          borderRadius: 6,
        },
        itemStyle: {
          borderRadius: 6,
          color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
            { offset: 0, color: barColor + 'cc' },
            { offset: 1, color: barColor },
          ]),
          shadowColor: barColor + '66',
          shadowBlur: 8,
        },
      },
    ],
  }
  chart.setOption(option)
  return chart.getDataURL({ backgroundColor })
}

/**
 * Glassmorphism KPI Card (twFathom Core Widget)
 * Generates an SVG DataURL with left accent bar, label, huge value + unit, and sparkline.
 */
export const kpi = (
  title: string,
  text: string,
  value: number,
  color: string,
  values: number[] = [],
  dark: boolean = true,
  w: number = 220,
  h: number = 84
): string => {
  const accentColor = color || COLOR_PRIMARY
  const safeWidth = Math.max(160, w)
  const safeHeight = Math.max(60, h)

  // Parse value and unit from formatted text or number
  let valStr = text || `${value.toFixed(1)}`
  let unitStr = ''
  
  // Extract trailing unit if present (e.g., "12.5 Mbps", "85 %", "24.5°C")
  const match = valStr.match(/^([0-9.,+-]+)\s*([a-zA-Z%°/]+.*)?$/)
  if (match) {
    valStr = match[1]
    unitStr = match[2] || ''
  }

  // Card theme colors
  const borderColor = dark ? 'rgba(255, 255, 255, 0.12)' : 'rgba(0, 0, 0, 0.1)'
  const titleColor = dark ? '#9ca3af' : '#64748b'
  const valueColor = dark ? '#f3f4f6' : '#0f172a'
  const unitColor = dark ? '#9ca3af' : '#64748b'

  // Sparkline calculation
  let sparklinePath = ''
  let sparklineAreaPath = ''
  if (values && values.length > 1) {
    const validValues = values.filter((v) => typeof v === 'number' && !isNaN(v) && isFinite(v))
    if (validValues.length > 1) {
      const minVal = Math.min(...validValues)
      const maxVal = Math.max(...validValues)
      const range = maxVal - minVal || 1
      const padX = 14
      const plotW = safeWidth - padX - 10
      const plotTop = safeHeight * 0.44
      const plotH = safeHeight - plotTop - 8

      const points: [number, number][] = validValues.map((v, i) => {
        const x = padX + (i / (validValues.length - 1)) * plotW
        const y = plotTop + plotH - ((v - minVal) / range) * plotH
        return [x, y]
      })

      sparklinePath = `M ${points.map(p => `${p[0].toFixed(1)},${p[1].toFixed(1)}`).join(' L ')}`
      sparklineAreaPath = `${sparklinePath} L ${points[points.length - 1][0].toFixed(1)},${(plotTop + plotH).toFixed(1)} L ${points[0][0].toFixed(1)},${(plotTop + plotH).toFixed(1)} Z`
    }
  }

  const uid = Math.random().toString(36).substring(2, 8)
  const cardGradId = `cardGrad_${uid}`
  const sparkAreaId = `sparkArea_${uid}`

  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${safeWidth}" height="${safeHeight}" viewBox="0 0 ${safeWidth} ${safeHeight}">
    <defs>
      <linearGradient id="${cardGradId}" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stop-color="${dark ? '#1a1d2d' : '#ffffff'}" stop-opacity="0.95" />
        <stop offset="100%" stop-color="${dark ? '#0c0e14' : '#f1f5f9'}" stop-opacity="0.9" />
      </linearGradient>
      <linearGradient id="${sparkAreaId}" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stop-color="${accentColor}" stop-opacity="0.32" />
        <stop offset="100%" stop-color="${accentColor}" stop-opacity="0.0" />
      </linearGradient>
    </defs>

    <!-- Card Background -->
    <rect x="0.5" y="0.5" width="${safeWidth - 1}" height="${safeHeight - 1}" rx="10" ry="10" fill="url(#${cardGradId})" stroke="${borderColor}" stroke-width="1" />

    <!-- Left Accent Bar -->
    <rect x="0" y="3" width="4" height="${safeHeight - 6}" rx="2" ry="2" fill="${accentColor}" />

    <!-- Title Label -->
    <text x="14" y="20" fill="${titleColor}" font-family="Outfit, Inter, -apple-system, sans-serif" font-size="10.5" font-weight="700" letter-spacing="0.05em">${escapeXml(title || 'METRIC')}</text>

    <!-- Sparkline (Background) -->
    ${sparklineAreaPath ? `<path d="${sparklineAreaPath}" fill="url(#${sparkAreaId})" />` : ''}
    ${sparklinePath ? `<path d="${sparklinePath}" fill="none" stroke="${accentColor}" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" opacity="0.8" />` : ''}

    <!-- Main Value + Unit -->
    <g transform="translate(14, ${safeHeight - 18})">
      <text fill="${valueColor}" font-family="Outfit, Inter, -apple-system, sans-serif" font-size="${Math.min(28, Math.max(16, safeHeight * 0.36))}" font-weight="800">${escapeXml(valStr)}<tspan dx="4" font-size="${Math.min(13, Math.max(10, safeHeight * 0.18))}" font-weight="500" fill="${unitColor}">${escapeXml(unitStr)}</tspan></text>
    </g>
  </svg>`

  return toSvgDataUrl(svg)
}

function toSvgDataUrl(svg: string): string {
  try {
    return 'data:image/svg+xml;base64,' + btoa(unescape(encodeURIComponent(svg)))
  } catch {
    return 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(svg)
  }
}

function escapeXml(unsafe: string): string {
  if (!unsafe) return ''
  return unsafe
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;')
}
