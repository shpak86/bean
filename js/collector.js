/**
 * Behavioral Metrics Collector
 */

class BehavioralMetricsCollector {
  constructor(options = {}) {
    this.metrics = {
      mouseMoves: 0,
      totalDistance: 0,
      moveIntervals: [],
      lastMouseMoveTime: null,
      clicks: 0,
      clickTimings: [],
      scrolls: 0,
      scrollTimings: [],
      textInputEvents: 0,
      textInputTimings: [],
      startTime: Date.now(),
      sessionDuration: 0,
      browser: this.parseBrowserInfo()
    };

    this.options = {
      autoStart: options.autoStart !== false,
      enableLogging: options.enableLogging || false,
      reportInterval: options.reportInterval || 5000, // 5 seconds
      address: options.address
    };

    this.lastClickTime = null;
    this.lastScrollTime = null;
    this.lastTextInputTime = null;

    if (this.options.autoStart) {
      this.start();
    }
  }

  /**
   * Parse browser information from userAgent
   */
  parseBrowserInfo() {
    const ua = navigator.userAgent;
    const browserInfo = {
      userAgent: ua,
      language: navigator.language,
      platform: navigator.platform,
      screenWidth: window.screen.width,
      screenHeight: window.screen.height,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      cookiesEnabled: navigator.cookieEnabled,
      onLine: navigator.onLine,
      deviceMemory: navigator.deviceMemory || 'unknown',
      maxTouchPoints: navigator.maxTouchPoints || 0,
      browser: this.detectBrowser(ua),
      os: this.detectOS(ua)
    };

    return browserInfo;
  }

  /**
   * Detect browser from userAgent
   */
  detectBrowser(ua) {
    const browsers = [
      { name: 'Chrome', pattern: /Chrome\/(\d+)/ },
      { name: 'Firefox', pattern: /Firefox\/(\d+)/ },
      { name: 'Safari', pattern: /Version\/(\d+).*Safari/ },
      { name: 'Edge', pattern: /Edg\/(\d+)/ },
      { name: 'Opera', pattern: /OPR\/(\d+)/ },
      { name: 'IE', pattern: /MSIE (\d+)|Trident.*rv:(\d+)/ }
    ];

    for (const browser of browsers) {
      const match = ua.match(browser.pattern);
      if (match) {
        return {
          name: browser.name,
          version: match[1] || match[2] || 'unknown'
        };
      }
    }

    return { name: 'Unknown', version: 'unknown' };
  }

  /**
   * Detect operating system from userAgent
   */
  detectOS(ua) {
    const systems = [
      { name: 'Windows', pattern: /Windows NT (\d+\.\d+)/ },
      { name: 'macOS', pattern: /Mac OS X (\d+[._]\d+)/ },
      { name: 'Linux', pattern: /Linux/ },
      { name: 'iOS', pattern: /iPhone OS (\d+[._]\d+)/ },
      { name: 'Android', pattern: /Android (\d+[._]\d+)/ }
    ];

    for (const system of systems) {
      const match = ua.match(system.pattern);
      if (match) {
        return {
          name: system.name,
          version: match[1] || 'unknown'
        };
      }
    }

    return { name: 'Unknown', version: 'unknown' };
  }

  /**
   * Start collecting metrics
   */
  start() {
    this.attachMouseMoveListener();
    this.attachClickListener();
    this.attachScrollListener();
    this.attachTextInputListener();
    this.attachVisibilityListener();

    if (this.options.reportInterval > 0) {
      this.startReportInterval();
    }

    this.log('Metrics collection started');
  }

  /**
   * Stop collecting metrics
   */
  stop() {
    this.removeMouseMoveListener();
    this.removeClickListener();
    this.removeScrollListener();
    this.removeTextInputListener();
    this.removeVisibilityListener();
    this.stopReportInterval();

    this.log('Metrics collection stopped');
  }

  /**
   * Attach mouse move listener (throttled)
   */
  attachMouseMoveListener() {
    let throttleTimer = null;
    const throttleDelay = 16;

    this.lastMousePosition = null;

    this.mouseMoveHandler = (event) => {
      if (!throttleTimer) {
        const now = Date.now();
        const currentPos = { x: event.clientX, y: event.clientY };

        if (this.lastMousePosition) {
          const dx = currentPos.x - this.lastMousePosition.x;
          const dy = currentPos.y - this.lastMousePosition.y;
          const distance = Math.sqrt(dx * dx + dy * dy);

          this.metrics.mouseMoves++;
          this.metrics.totalDistance += distance;

          const timeDelta = now - this.lastMouseMoveTime;
          if (timeDelta > 0) {
            this.metrics.moveIntervals.push(timeDelta);
          }
        }

        this.lastMousePosition = currentPos;
        this.lastMouseMoveTime = now;

        throttleTimer = setTimeout(() => {
          throttleTimer = null;
        }, throttleDelay);
      }
    };

    document.addEventListener('mousemove', this.mouseMoveHandler, { passive: true });
  }

  /**
   * Remove mouse move listener
   */
  removeMouseMoveListener() {
    if (this.mouseMoveHandler) {
      document.removeEventListener('mousemove', this.mouseMoveHandler);
    }
  }

  /**
   * Attach click listener
   */
  attachClickListener() {
    this.clickHandler = (event) => {
      const now = Date.now();

      this.metrics.clicks++;

      if (this.lastClickTime !== null) {
        const timingDiff = now - this.lastClickTime;
        this.metrics.clickTimings.push(timingDiff);
      }

      this.lastClickTime = now;
      this.log(`Click detected. Total: ${this.metrics.clicks}`);
    };

    document.addEventListener('click', this.clickHandler, true);
  }

  /**
   * Remove click listener
   */
  removeClickListener() {
    if (this.clickHandler) {
      document.removeEventListener('click', this.clickHandler, true);
    }
  }

  /**
   * Attach scroll listener (throttled)
   */
  attachScrollListener() {
    let throttleTimer = null;
    const throttleDelay = 20; // Throttle scroll events to every 20ms

    this.scrollHandler = () => {
      if (!throttleTimer) {
        const now = Date.now();

        this.metrics.scrolls++;

        if (this.lastScrollTime !== null) {
          const timingDiff = now - this.lastScrollTime;
          this.metrics.scrollTimings.push(timingDiff);
        }

        this.lastScrollTime = now;
        this.log(`Scroll detected. Total: ${this.metrics.scrolls}`);

        throttleTimer = setTimeout(() => {
          throttleTimer = null;
        }, throttleDelay);
      }
    };

    document.addEventListener('scroll', this.scrollHandler, { passive: true });
  }

  /**
   * Remove scroll listener
   */
  removeScrollListener() {
    if (this.scrollHandler) {
      document.removeEventListener('scroll', this.scrollHandler);
    }
  }

  /**
   * Attach text input listener
   */
  attachTextInputListener() {
    this.textInputHandler = (event) => {
      // Track input, textarea, contenteditable elements
      const target = event.target;
      const isTextInput =
        target.tagName === 'INPUT' &&
        ['text', 'email', 'password', 'search', 'url', 'tel'].includes(target.type);
      const isTextarea = target.tagName === 'TEXTAREA';
      const isContentEditable = target.contentEditable === 'true';

      if (isTextInput || isTextarea || isContentEditable) {
        const now = Date.now();

        this.metrics.textInputEvents++;

        if (this.lastTextInputTime !== null) {
          const timingDiff = now - this.lastTextInputTime;
          this.metrics.textInputTimings.push(timingDiff);
        }

        this.lastTextInputTime = now;
        this.log(`Text input detected. Total: ${this.metrics.textInputEvents}`);
      }
    };

    document.addEventListener('input', this.textInputHandler, true);
  }

  /**
   * Remove text input listener
   */
  removeTextInputListener() {
    if (this.textInputHandler) {
      document.removeEventListener('input', this.textInputHandler, true);
    }
  }

  /**
   * Attach visibility listener to track session duration
   */
  attachVisibilityListener() {
    this.visibilityHandler = () => {
      if (document.visibilityState === 'hidden') {
        this.metrics.sessionDuration = Date.now() - this.metrics.startTime;
        this.log('Page hidden, session duration updated');
      }
    };

    document.addEventListener('visibilitychange', this.visibilityHandler);
  }

  /**
   * Remove visibility listener
   */
  removeVisibilityListener() {
    if (this.visibilityHandler) {
      document.removeEventListener('visibilitychange', this.visibilityHandler);
    }
  }

  /**
   * Start automatic reporting
   */
  startReportInterval() {
    this.reportIntervalId = setInterval(() => {
      // this.report();
      this.sendToServer(this.options.address);
    }, this.options.reportInterval);
  }

  /**
   * Stop automatic reporting
   */
  stopReportInterval() {
    if (this.reportIntervalId) {
      clearInterval(this.reportIntervalId);
      this.reportIntervalId = null;
    }
  }

  /**
   * Get calculated metrics
   */
  getCalculatedMetrics() {
    const calculated = {};

    // Moves timings analysis
    if (this.metrics.moveIntervals.length > 0) {
      const avgInterval = this.metrics.moveIntervals.reduce((a, b) => a + b, 0) / this.metrics.moveIntervals.length;
      const avgSpeed = this.metrics.totalDistance / (this.metrics.moveIntervals.reduce((a, b) => a + b, 0) / 1000); // px/sec

      calculated.mouseStats = {
        totalDistance: Math.round(this.metrics.totalDistance),
        moveCount: this.metrics.moveIntervals.length,
        avgInterval: Math.round(avgInterval),
        minInterval: Math.min(...this.metrics.moveIntervals),
        maxInterval: Math.max(...this.metrics.moveIntervals),
        avgSpeed: +avgSpeed.toFixed(2)
      };
    } else {
      calculated.mouseStats = {
        totalDistance: 0,
        moveCount: 0,
        avgInterval: 0,
        minInterval: 0,
        maxInterval: 0,
        avgSpeed: 0
      };
    }

    // Click timings analysis
    if (this.metrics.clickTimings.length > 0) {
      calculated.clickTimingStats = {
        min: Math.min(...this.metrics.clickTimings),
        max: Math.max(...this.metrics.clickTimings),
        avg: Math.round(
          this.metrics.clickTimings.reduce((a, b) => a + b, 0) /
          this.metrics.clickTimings.length
        ),
        count: this.metrics.clickTimings.length
      };
    } else {
      calculated.clickTimingStats = {
        min: 0,
        max: 0,
        avg: 0,
        count: 0
      };
    }

    // Scroll timings analysis
    if (this.metrics.scrollTimings.length > 0) {
      calculated.scrollTimingStats = {
        min: Math.min(...this.metrics.scrollTimings),
        max: Math.max(...this.metrics.scrollTimings),
        avg: Math.round(
          this.metrics.scrollTimings.reduce((a, b) => a + b, 0) /
          this.metrics.scrollTimings.length
        ),
        count: this.metrics.scrollTimings.length
      };
    } else {
      calculated.scrollTimingStats = {
        min: 0,
        max: 0,
        avg: 0,
        count: 0
      };
    }

    // Text input timings analysis
    if (this.metrics.textInputTimings.length > 0) {
      calculated.textInputTimingStats = {
        min: Math.min(...this.metrics.textInputTimings),
        max: Math.max(...this.metrics.textInputTimings),
        avg: Math.round(
          this.metrics.textInputTimings.reduce((a, b) => a + b, 0) /
          this.metrics.textInputTimings.length
        ),
        count: this.metrics.textInputTimings.length
      };
    } else {
      calculated.textInputTimingStats = {
        min: 0,
        max: 0,
        avg: 0,
        count: 0
      };
    }

    // Session duration
    calculated.sessionDuration =
      this.metrics.sessionDuration || Date.now() - this.metrics.startTime;

    return calculated;
  }

  /**
   * Get all metrics including calculated ones
   */
  getMetrics() {
    return {
      ...this.metrics,
      calculated: this.getCalculatedMetrics()
    };
  }

  /**
   * Generate a report
   */
  report() {
    const report = this.getMetrics();
    if (this.options.address && (report.mouseMoves || report.clicks || report.scrolls)) {
      this.log('Report sent:', report);
      // this.sendToServer(this.options.address);
    }
    this.reset();
  }

  /**
   * Reset all metrics
   */
  reset() {
    this.metrics = {
      mouseMoves: 0,
      totalDistance: 0,
      moveIntervals: [],
      clicks: 0,
      clickTimings: [],
      scrolls: 0,
      scrollTimings: [],
      textInputEvents: 0,
      textInputTimings: [],
      startTime: Date.now(),
      sessionDuration: 0,
      browser: this.metrics.browser
    };

    this.lastClickTime = null;
    this.lastScrollTime = null;
    this.lastTextInputTime = null;
    this.lastMousePosition = null;
    this.lastMouseMoveTime = null;

    this.log('Metrics reset');
  }

  /**
   * Logging utility
   */
  log(message, data = null) {
    if (this.options.enableLogging) {
      console.log(`[BehavioralMetrics] ${message}`, data || '');
    }
  }

  /**
   * Send metrics to server as a flat object
   */
  sendToServer(url) {
    const metrics = this.getMetrics();
    const calculated = metrics.calculated;

    // Extract browser and OS details
    const browserInfo = metrics.browser;
    const browser = browserInfo.browser || { name: 'Unknown', version: 'unknown' };
    const os = browserInfo.os || { name: 'Unknown', version: 'unknown' };

    // Normalize deviceMemory: convert to number, default to 0 if 'unknown'
    const deviceMemory = typeof browserInfo.deviceMemory === 'number'
      ? browserInfo.deviceMemory
      : 0;

    // Prepare flat payload
    const payload = {
      // Behavior Metrics
      mouseMoves: metrics.mouseMoves,
      clicks: metrics.clicks,
      clickTimingMin: calculated.clickTimingStats.min,
      clickTimingMax: calculated.clickTimingStats.max,
      clickTimingAvg: calculated.clickTimingStats.avg,
      clickTimingCount: calculated.clickTimingStats.count,
      scrolls: metrics.scrolls,
      scrollTimingMin: calculated.scrollTimingStats.min,
      scrollTimingMax: calculated.scrollTimingStats.max,
      scrollTimingAvg: calculated.scrollTimingStats.avg,
      scrollTimingCount: calculated.scrollTimingStats.count,
      textInputEvents: metrics.textInputEvents,
      textInputTimingMin: calculated.textInputTimingStats.min,
      textInputTimingMax: calculated.textInputTimingStats.max,
      textInputTimingAvg: calculated.textInputTimingStats.avg,
      textInputTimingCount: calculated.textInputTimingStats.count,
      sessionDuration: calculated.sessionDuration,

      // Timestamp
      timestamp: new Date().toISOString(),

      // Browser and Device Info
      userAgent: browserInfo.userAgent,
      language: browserInfo.language,
      platform: browserInfo.platform,
      screenWidth: browserInfo.screenWidth,
      screenHeight: browserInfo.screenHeight,
      timezone: browserInfo.timezone,
      cookiesEnabled: browserInfo.cookiesEnabled,
      onLine: browserInfo.onLine,
      deviceMemory: deviceMemory,
      maxTouchPoints: browserInfo.maxTouchPoints,

      // Browser Details
      browserName: browser.name,
      browserVersion: browser.version,

      // OS Details
      osName: os.name,
      osVersion: os.version
    };

    fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(payload),
      keepalive: true
    })
      .then(response => {
        if (response.ok) {
          this.log('Metrics sent successfully');
        } else {
          this.log('Failed to send metrics. Status:', response.status);
        }
      })
      .catch(error => {
        this.log('Error sending metrics:', error);
      });

      this.reset();
  }

}

// Export for use as module or standalone
if (typeof module !== 'undefined' && module.exports) {
  module.exports = BehavioralMetricsCollector;
}