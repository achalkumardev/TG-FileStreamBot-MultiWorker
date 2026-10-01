package routes

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/bot"
	"EverythingSuckz/fsb/internal/types"
	"EverythingSuckz/fsb/internal/utils"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
)

func (e *allRoutes) LoadWatch(r *Route) {
	r.Engine.GET("/watch/:messageID", getWatchRoute)
	r.Engine.HEAD("/watch/:messageID", getWatchRoute)
}

type watchData struct {
	FileName    string
	FileSize    string
	MimeType    string
	StreamURL   string
	DownloadURL string
	Host        string
}

const watchHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
  <title>{{ .FileName }} • High-Speed Stream & Download</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700;800&display=swap" rel="stylesheet">
  <link rel="stylesheet" href="https://cdn.plyr.io/3.7.8/plyr.css" />
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css" />
  <style>
    :root {
      --bg-dark: #07090e;
      --card-bg: rgba(15, 20, 32, 0.72);
      --card-border: rgba(255, 255, 255, 0.08);
      --card-border-hover: rgba(99, 102, 241, 0.4);
      --primary: #6366f1;
      --primary-glow: rgba(99, 102, 241, 0.35);
      --emerald: #10b981;
      --text-main: #f8fafc;
      --text-secondary: #94a3b8;
      --text-tertiary: #64748b;
    }

    * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
      font-family: 'Plus Jakarta Sans', -apple-system, BlinkMacSystemFont, sans-serif;
      -webkit-tap-highlight-color: transparent;
    }

    body {
      background: radial-gradient(circle at 50% -20%, #1e1b4b 0%, #0c101c 45%, #07090e 100%);
      background-attachment: fixed;
      color: var(--text-main);
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: 16px 14px 48px;
    }

    /* Ambient animated background orbs */
    .ambient-bg {
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      pointer-events: none;
      z-index: 0;
      overflow: hidden;
    }
    .orb {
      position: absolute;
      border-radius: 50%;
      filter: blur(90px);
      opacity: 0.18;
      animation: floatOrb 12s ease-in-out infinite alternate;
    }
    .orb-1 {
      width: 380px;
      height: 380px;
      background: #6366f1;
      top: -80px;
      right: 10%;
    }
    .orb-2 {
      width: 320px;
      height: 320px;
      background: #06b6d4;
      bottom: 20%;
      left: 5%;
      animation-delay: -5s;
    }
    @keyframes floatOrb {
      0% { transform: translateY(0px) scale(1); }
      100% { transform: translateY(35px) scale(1.08); }
    }

    .container {
      position: relative;
      z-index: 1;
      width: 100%;
      max-width: 860px;
      display: flex;
      flex-direction: column;
      gap: 18px;
    }

    /* Top Navigation / Brand */
    .top-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 6px 4px;
    }
    .brand-logo {
      display: flex;
      align-items: center;
      gap: 10px;
      font-weight: 800;
      font-size: 1.12rem;
      letter-spacing: -0.02em;
      background: linear-gradient(135deg, #ffffff 0%, #cbd5e1 50%, #818cf8 100%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
    }
    .brand-icon-wrap {
      width: 34px;
      height: 34px;
      border-radius: 10px;
      background: linear-gradient(135deg, #4f46e5, #7c3aed);
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
      font-size: 0.95rem;
      box-shadow: 0 4px 14px rgba(99, 102, 241, 0.4);
      -webkit-text-fill-color: initial;
    }
    .server-status {
      display: flex;
      align-items: center;
      gap: 7px;
      font-size: 0.78rem;
      font-weight: 600;
      padding: 5px 12px;
      border-radius: 9999px;
      background: rgba(16, 185, 129, 0.1);
      border: 1px solid rgba(16, 185, 129, 0.25);
      color: #34d399;
    }
    .status-dot {
      width: 7px;
      height: 7px;
      background: #10b981;
      border-radius: 50%;
      box-shadow: 0 0 10px #10b981, 0 0 4px #10b981;
      animation: pulseDot 2s infinite;
    }
    @keyframes pulseDot {
      0%, 100% { opacity: 1; transform: scale(1); }
      50% { opacity: 0.5; transform: scale(0.85); }
    }

    /* Player Container with Cinema Glow */
    .player-card {
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 22px;
      overflow: hidden;
      box-shadow: 0 20px 50px -15px rgba(0, 0, 0, 0.8), 0 0 40px -15px var(--primary-glow);
      backdrop-filter: blur(20px);
      -webkit-backdrop-filter: blur(20px);
      transition: border-color 0.3s ease;
    }
    .player-card:hover {
      border-color: rgba(255, 255, 255, 0.14);
    }
    .player-wrapper {
      position: relative;
      width: 100%;
      background: #000;
      aspect-ratio: 16 / 9;
      max-height: 500px;
    }
    .plyr {
      height: 100%;
      width: 100%;
      --plyr-color-main: #6366f1;
      --plyr-video-background: #000;
      --plyr-badge-text-color: #fff;
    }

    /* Video Details */
    .video-info-section {
      padding: 20px 22px;
      border-top: 1px solid var(--card-border);
      background: linear-gradient(180deg, rgba(255,255,255,0.015) 0%, rgba(255,255,255,0) 100%);
    }
    .file-name {
      font-size: 1.18rem;
      font-weight: 700;
      line-height: 1.45;
      color: #f1f5f9;
      margin-bottom: 12px;
      word-break: break-word;
      letter-spacing: -0.01em;
    }
    .specs-grid {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }
    .spec-chip {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      font-size: 0.8rem;
      font-weight: 600;
      padding: 6px 12px;
      border-radius: 10px;
      background: rgba(255, 255, 255, 0.04);
      border: 1px solid rgba(255, 255, 255, 0.07);
      color: var(--text-secondary);
      transition: all 0.2s ease;
    }
    .spec-chip i {
      color: #818cf8;
      font-size: 0.85rem;
    }
    .spec-chip.emerald i {
      color: #34d399;
    }
    .spec-chip.cyan i {
      color: #22d3ee;
    }

    /* Section Heading */
    .section-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin: 6px 2px 2px;
    }
    .section-title {
      font-size: 0.86rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: var(--text-secondary);
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .section-title i {
      color: var(--primary);
    }

    /* Grid of App Action Cards */
    .action-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 12px;
    }
    @media (min-width: 680px) {
      .action-grid {
        grid-template-columns: repeat(3, 1fr);
        gap: 14px;
      }
    }

    /* App Card Styling */
    .app-card {
      position: relative;
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 18px;
      padding: 14px 15px;
      display: flex;
      align-items: center;
      gap: 12px;
      cursor: pointer;
      text-align: left;
      text-decoration: none;
      color: inherit;
      backdrop-filter: blur(14px);
      -webkit-backdrop-filter: blur(14px);
      transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
      box-shadow: 0 4px 15px rgba(0, 0, 0, 0.25);
      overflow: hidden;
    }
    .app-card::before {
      content: '';
      position: absolute;
      inset: 0;
      border-radius: 18px;
      padding: 1px;
      background: linear-gradient(135deg, rgba(255,255,255,0.12), rgba(255,255,255,0));
      -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
      -webkit-mask-composite: xor;
      mask-composite: exclude;
      pointer-events: none;
    }
    .app-card:hover {
      transform: translateY(-2px);
      border-color: var(--card-border-hover);
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4), 0 0 20px -5px var(--hover-glow, rgba(99, 102, 241, 0.25));
    }
    .app-card:active {
      transform: scale(0.97);
    }

    .app-icon {
      width: 44px;
      height: 44px;
      border-radius: 13px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 1.3rem;
      flex-shrink: 0;
      color: #fff;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    }
    .app-details {
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
    .app-name {
      font-size: 0.94rem;
      font-weight: 700;
      color: #fff;
      letter-spacing: -0.01em;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .app-sub {
      font-size: 0.72rem;
      color: var(--text-secondary);
      font-weight: 500;
      margin-top: 2px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    /* Specific App Themes */
    .card-vlc { --hover-glow: rgba(249, 115, 22, 0.3); }
    .card-vlc .app-icon { background: linear-gradient(135deg, #f97316, #ea580c); }

    .card-mx { --hover-glow: rgba(59, 130, 246, 0.3); }
    .card-mx .app-icon { background: linear-gradient(135deg, #3b82f6, #1d4ed8); }

    .card-playit { --hover-glow: rgba(168, 85, 247, 0.3); }
    .card-playit .app-icon { background: linear-gradient(135deg, #a855f7, #7c3aed); }

    .card-1dm { --hover-glow: rgba(6, 182, 212, 0.4); }
    .card-1dm .app-icon { background: linear-gradient(135deg, #06b6d4, #0284c7); }

    .card-other { --hover-glow: rgba(148, 163, 184, 0.2); }
    .card-other .app-icon { background: linear-gradient(135deg, #475569, #334155); }

    .card-copy { --hover-glow: rgba(99, 102, 241, 0.3); }
    .card-copy .app-icon { background: linear-gradient(135deg, #4f46e5, #4338ca); }

    /* Hero Direct Download Banner */
    .hero-download-btn {
      position: relative;
      background: linear-gradient(135deg, #059669 0%, #10b981 50%, #047857 100%);
      border: 1px solid rgba(52, 211, 153, 0.3);
      border-radius: 18px;
      padding: 16px 22px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      color: #fff;
      text-decoration: none;
      cursor: pointer;
      box-shadow: 0 10px 25px -5px rgba(16, 185, 129, 0.4), 0 0 20px rgba(16, 185, 129, 0.2);
      transition: all 0.25s ease;
      overflow: hidden;
    }
    .hero-download-btn:hover {
      transform: translateY(-2px);
      box-shadow: 0 14px 32px -5px rgba(16, 185, 129, 0.5), 0 0 25px rgba(16, 185, 129, 0.3);
    }
    .hero-download-btn:active {
      transform: scale(0.98);
    }
    .hero-dl-left {
      display: flex;
      align-items: center;
      gap: 14px;
    }
    .hero-dl-icon {
      width: 46px;
      height: 46px;
      border-radius: 12px;
      background: rgba(255, 255, 255, 0.18);
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 1.4rem;
    }
    .hero-dl-title {
      font-size: 1.05rem;
      font-weight: 800;
      letter-spacing: -0.01em;
    }
    .hero-dl-sub {
      font-size: 0.78rem;
      color: rgba(255, 255, 255, 0.85);
      font-weight: 500;
      margin-top: 2px;
    }
    .hero-dl-arrow {
      font-size: 1.2rem;
      opacity: 0.85;
      padding-right: 4px;
    }

    /* Instructions & Speed Tips Collapsible Card */
    .tips-card {
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 18px;
      padding: 16px 18px;
      backdrop-filter: blur(14px);
      -webkit-backdrop-filter: blur(14px);
    }
    .tips-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      cursor: pointer;
      user-select: none;
    }
    .tips-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 0.88rem;
      font-weight: 700;
      color: #cbd5e1;
    }
    .tips-title i {
      color: #38bdf8;
    }
    .tips-toggle-icon {
      color: var(--text-tertiary);
      transition: transform 0.25s ease;
      font-size: 0.85rem;
    }
    .tips-toggle-icon.open {
      transform: rotate(180deg);
    }
    .tips-content {
      display: none;
      margin-top: 14px;
      padding-top: 14px;
      border-top: 1px solid var(--card-border);
      font-size: 0.82rem;
      color: var(--text-secondary);
      line-height: 1.6;
    }
    .tips-content.show {
      display: block;
    }
    .tip-step {
      display: flex;
      align-items: flex-start;
      gap: 10px;
      margin-bottom: 10px;
    }
    .tip-step:last-child {
      margin-bottom: 0;
    }
    .step-badge {
      width: 20px;
      height: 20px;
      border-radius: 6px;
      background: rgba(99, 102, 241, 0.15);
      border: 1px solid rgba(99, 102, 241, 0.3);
      color: #a5b4fc;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 0.72rem;
      font-weight: 700;
      flex-shrink: 0;
      margin-top: 1px;
    }

    /* Footer */
    .footer {
      text-align: center;
      font-size: 0.75rem;
      color: var(--text-tertiary);
      margin-top: 4px;
    }

    /* Toast Notification */
    .toast {
      position: fixed;
      bottom: 24px;
      left: 50%;
      transform: translateX(-50%) translateY(90px);
      background: rgba(15, 23, 42, 0.95);
      border: 1px solid rgba(99, 102, 241, 0.4);
      color: #fff;
      padding: 12px 22px;
      border-radius: 9999px;
      font-size: 0.88rem;
      font-weight: 600;
      display: flex;
      align-items: center;
      gap: 10px;
      box-shadow: 0 15px 35px rgba(0, 0, 0, 0.6), 0 0 20px rgba(99, 102, 241, 0.25);
      opacity: 0;
      pointer-events: none;
      transition: all 0.32s cubic-bezier(0.16, 1, 0.3, 1);
      z-index: 9999;
      backdrop-filter: blur(12px);
    }
    .toast.show {
      transform: translateX(-50%) translateY(0);
      opacity: 1;
      pointer-events: auto;
    }
    .toast i {
      color: #34d399;
      font-size: 1.05rem;
    }
  </style>
</head>
<body>
  <!-- Ambient background glow -->
  <div class="ambient-bg">
    <div class="orb orb-1"></div>
    <div class="orb orb-2"></div>
  </div>

  <div class="container">
    <!-- Top Header -->
    <div class="top-header">
      <div class="brand-logo">
        <div class="brand-icon-wrap">
          <i class="fa-solid fa-play"></i>
        </div>
        <span>StreamEngine Pro</span>
      </div>
      <div class="server-status">
        <span class="status-dot"></span> 10Gbps Live
      </div>
    </div>

    <!-- Video Player Card -->
    <div class="player-card">
      <div class="player-wrapper">
        <video id="player" playsinline controls preload="metadata">
          <source src="{{ .StreamURL }}" type="{{ .MimeType }}">
          Your browser does not support HTML5 video streaming.
        </video>
      </div>

      <!-- File Details & Specs -->
      <div class="video-info-section">
        <h1 class="file-name">{{ .FileName }}</h1>
        <div class="specs-grid">
          <span class="spec-chip emerald"><i class="fa-solid fa-hard-drive"></i> {{ .FileSize }}</span>
          <span class="spec-chip cyan"><i class="fa-solid fa-video"></i> {{ .MimeType }}</span>
          <span class="spec-chip"><i class="fa-solid fa-bolt"></i> High Concurrency</span>
          <span class="spec-chip"><i class="fa-solid fa-shield-halved"></i> SSL Encrypted</span>
        </div>
      </div>
    </div>

    <!-- Direct Fast Download Hero Button -->
    <a class="hero-download-btn" href="{{ .DownloadURL }}">
      <div class="hero-dl-left">
        <div class="hero-dl-icon">
          <i class="fa-solid fa-cloud-arrow-down"></i>
        </div>
        <div>
          <div class="hero-dl-title">Direct High-Speed Download</div>
          <div class="hero-dl-sub">Full Jio Speed • Resume Supported • Zero Disk Limit</div>
        </div>
      </div>
      <div class="hero-dl-arrow">
        <i class="fa-solid fa-chevron-right"></i>
      </div>
    </a>

    <!-- Action Section Header -->
    <div class="section-header">
      <div class="section-title">
        <i class="fa-solid fa-rocket"></i> Launch in External App
      </div>
    </div>

    <!-- Grid Action Cards -->
    <div class="action-grid">
      <!-- VLC Player -->
      <div class="app-card card-vlc" onclick="openVLC()">
        <div class="app-icon">
          <i class="fa-brands fa-vimeo-v"></i>
        </div>
        <div class="app-details">
          <div class="app-name">VLC Player</div>
          <div class="app-sub">Dual-Audio & Subtitles</div>
        </div>
      </div>

      <!-- MX Player -->
      <div class="app-card card-mx" onclick="openMX()">
        <div class="app-icon">
          <i class="fa-solid fa-play"></i>
        </div>
        <div class="app-details">
          <div class="app-name">MX Player</div>
          <div class="app-sub">HW+ Acceleration</div>
        </div>
      </div>

      <!-- PlayIt App -->
      <div class="app-card card-playit" onclick="openPlayIt()">
        <div class="app-icon">
          <i class="fa-solid fa-circle-play"></i>
        </div>
        <div class="app-details">
          <div class="app-name">PlayIt App</div>
          <div class="app-sub">Floating HD Player</div>
        </div>
      </div>

      <!-- 1DM Downloader -->
      <div class="app-card card-1dm" onclick="open1DM()">
        <div class="app-icon">
          <i class="fa-solid fa-bolt-lightning"></i>
        </div>
        <div class="app-details">
          <div class="app-name">1DM+ Downloader</div>
          <div class="app-sub">Up to 32 Threads Max DL</div>
        </div>
      </div>

      <!-- Other Players -->
      <div class="app-card card-other" onclick="openChooser()">
        <div class="app-icon">
          <i class="fa-solid fa-ellipsis"></i>
        </div>
        <div class="app-details">
          <div class="app-name">Other Players</div>
          <div class="app-sub">System App Chooser</div>
        </div>
      </div>

      <!-- Copy Link -->
      <div class="app-card card-copy" onclick="copyLink()">
        <div class="app-icon">
          <i class="fa-regular fa-copy"></i>
        </div>
        <div class="app-details">
          <div class="app-name">Copy Link</div>
          <div class="app-sub">Paste in any player/app</div>
        </div>
      </div>
    </div>

    <!-- Speed & Audio Guide -->
    <div class="tips-card">
      <div class="tips-header" onclick="toggleTips()">
        <div class="tips-title">
          <i class="fa-solid fa-circle-info"></i>
          <span>Tips for Jio 50MB/s Speed & Audio Tracks</span>
        </div>
        <i id="tips-icon" class="fa-solid fa-chevron-down tips-toggle-icon"></i>
      </div>
      <div id="tips-body" class="tips-content">
        <div class="tip-step">
          <span class="step-badge">1</span>
          <div><strong>1DM Max Speed:</strong> In 1DM App, set <em>Number of parts = 16 or 32</em> in download settings. This bypasses Jio single-stream limits and gives full 50MB/s speed.</div>
        </div>
        <div class="tip-step">
          <span class="step-badge">2</span>
          <div><strong>Audio & Subtitle Switching:</strong> For MKV files with multi-language audio, open in VLC or MX Player and tap the Audio/Subtitle track icon to switch between Hindi/English.</div>
        </div>
        <div class="tip-step">
          <span class="step-badge">3</span>
          <div><strong>Direct Streaming:</strong> You can watch immediately without waiting for download. Seeking is instant via VPS range-streaming.</div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="footer">
      Powered by High-Performance Go Stream Engine & Cloudflare Edge
    </div>
  </div>

  <!-- Toast Notification -->
  <div id="toast" class="toast">
    <i class="fa-solid fa-circle-check"></i> <span id="toast-msg">Stream Link Copied!</span>
  </div>

  <script src="https://cdn.plyr.io/3.7.8/plyr.polyfilled.js"></script>
  <script>
    const streamUrl = "{{ .StreamURL }}";
    const downloadUrl = "{{ .DownloadURL }}";
    const fileName = "{{ .FileName }}";

    // Initialize Plyr Player with custom controls
    const player = new Plyr('#player', {
      controls: [
        'play-large', 'restart', 'rewind', 'play', 'fast-forward',
        'progress', 'current-time', 'duration', 'mute', 'volume',
        'settings', 'pip', 'airplay', 'fullscreen'
      ],
      seekTime: 10,
      keyboard: { focused: true, global: true },
      tooltips: { controls: true, seek: true }
    });

    function toggleTips() {
      const body = document.getElementById('tips-body');
      const icon = document.getElementById('tips-icon');
      body.classList.toggle('show');
      icon.classList.toggle('open');
    }

    function showToast(text) {
      const toast = document.getElementById('toast');
      document.getElementById('toast-msg').innerText = text;
      toast.classList.add('show');
      setTimeout(() => toast.classList.remove('show'), 2800);
    }

    function copyLink() {
      if (navigator.clipboard) {
        navigator.clipboard.writeText(streamUrl).then(() => {
          showToast("Stream Link Copied! Paste in VLC, MX, or 1DM");
        }).catch(() => {
          fallbackCopy();
        });
      } else {
        fallbackCopy();
      }
    }

    function fallbackCopy() {
      const ta = document.createElement("textarea");
      ta.value = streamUrl;
      document.body.appendChild(ta);
      ta.select();
      document.execCommand("copy");
      document.body.removeChild(ta);
      showToast("Stream Link Copied!");
    }

    function getCleanUrl() {
      return streamUrl.replace(/^https?:\/\//, '');
    }

    function openVLC() {
      const isAndroid = /android/i.test(navigator.userAgent);
      if (isAndroid) {
        const vlcIntent = "intent://" + getCleanUrl() + "#Intent;action=android.intent.action.VIEW;scheme=https;type=video/*;package=org.videolan.vlc;S.title=" + encodeURIComponent(fileName) + ";end";
        window.location.href = vlcIntent;
      } else {
        window.location.href = "vlc://" + streamUrl;
      }
    }

    function openMX() {
      const mxIntent = "intent://" + getCleanUrl() + "#Intent;action=android.intent.action.VIEW;scheme=https;type=video/*;package=com.mxtech.videoplayer.ad;S.title=" + encodeURIComponent(fileName) + ";end";
      window.location.href = mxIntent;
    }

    function openPlayIt() {
      const playitIntent = "intent://" + getCleanUrl() + "#Intent;action=android.intent.action.VIEW;scheme=https;type=video/*;package=com.playit.videoplayer;S.title=" + encodeURIComponent(fileName) + ";end";
      window.location.href = playitIntent;
    }

    function open1DM() {
      const isAndroid = /android/i.test(navigator.userAgent);
      if (isAndroid) {
        const cleanDownload = downloadUrl.replace(/^https?:\/\//, '');
        const idmIntent = "intent://" + cleanDownload + "#Intent;action=android.intent.action.VIEW;scheme=https;type=video/*;package=idm.internet.download.manager;end";
        window.location.href = idmIntent;
        setTimeout(() => {
          showToast("Opening 1DM... You can also copy link & paste in 1DM");
        }, 1500);
      } else {
        window.location.href = downloadUrl;
      }
    }

    function openChooser() {
      const chooserIntent = "intent://" + getCleanUrl() + "#Intent;action=android.intent.action.VIEW;scheme=https;type=video/*;S.title=" + encodeURIComponent(fileName) + ";end";
      window.location.href = chooserIntent;
    }
  </script>
</body>
</html>`

func getWatchRoute(ctx *gin.Context) {
	messageIDParm := ctx.Param("messageID")
	messageID, err := strconv.Atoi(messageIDParm)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Invalid message ID")
		return
	}

	authHash := ctx.Query("hash")
	if authHash == "" {
		ctx.String(http.StatusBadRequest, "Missing hash parameter")
		return
	}

	client := bot.DefaultClient
	if client == nil {
		client = bot.GetNextWorker().Client
	}
	file, err := utils.TimeFuncWithResult(log, "FileFromMessage", func() (*types.File, error) {
		return utils.FileFromMessage(ctx, client, messageID)
	})
	if err != nil {
		ctx.String(http.StatusNotFound, "File not found")
		return
	}

	expectedHash := utils.PackFile(
		file.FileName,
		file.FileSize,
		file.MimeType,
		file.ID,
	)
	if !utils.CheckHash(authHash, expectedHash) {
		ctx.String(http.StatusUnauthorized, "Invalid hash")
		return
	}

	currentHost := config.ValueOf.Host
	if reqHost := ctx.Request.Host; reqHost != "" {
		scheme := "https"
		if proto := ctx.GetHeader("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		}
		currentHost = fmt.Sprintf("%s://%s", scheme, reqHost)
	}

	streamURL := fmt.Sprintf("%s/stream/%d/%s?hash=%s", currentHost, messageID, url.PathEscape(file.FileName), authHash)
	downloadURL := streamURL + "&d=true"

	tmpl, err := template.New("watch").Parse(watchHTML)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Template parsing error")
		return
	}

	data := watchData{
		FileName:    file.FileName,
		FileSize:    humanize.Bytes(uint64(file.FileSize)),
		MimeType:    file.MimeType,
		StreamURL:   streamURL,
		DownloadURL: downloadURL,
		Host:        currentHost,
	}

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Status(http.StatusOK)
	if err := tmpl.Execute(ctx.Writer, data); err != nil {
		log.Sugar().Errorf("Failed to execute template: %v", err)
	}
}
