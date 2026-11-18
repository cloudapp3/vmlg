const I18N = {
    'zh-CN': {
        your_ip: '你的 IP',
        probe_me: '测我本机',
        nodes_title: '可用测试节点',
        refresh: '刷新状态',
        tab_compare: '⚡ 多节点同屏横评',
        tab_diag: '🧭 路由追踪 & Ping',
        tab_speed: '🚀 带宽测速 & iPerf3',
        tab_bgp: '🌐 BGP 路由 & 拓扑',
        tab_share: '📋 脱敏报告导出',
        target_label: '目标 IP 或域名',
        ip_version: '协议栈',
        diag_tool: '诊断工具',
        start_test: '开始测试',
        stop_test: '停止',
        compare_heading: '全球节点并发延迟与回程线路对比',
        compare_desc: '一次输入访客 IP，所有在线节点同步探测延迟、丢包率并自动识别回程骨干网类型（CN2 GIA / 9929 / CMIN2 / IXP）。',
        include_route: '同步追踪回程线路',
        col_node: '节点',
        col_route: '回程线路判定',
        col_avg_rtt: '平均延迟',
        col_minmax: '最值 / 抖动',
        col_loss: '丢包率',
        col_detail: '路径详情',
        compare_empty: '点击上方「测我本机」或输入目标后点击「开始测试」，即可同屏对比全部在线节点。',
        diag_current_node: '当前节点：',
        diag_ready: '选择上方节点并启动实时流式 MTR 或 Ping 诊断。',
        stat_avg: '平均 RTT',
        stat_min_max: '最小 / 最大',
        stat_jitter: '抖动 (Jitter)',
        stat_loss: '丢包率',
        detected_route: '识别线路类型：',
        col_hop_host: '跳点 IP / PTR',
        col_asn_org: 'ASN / 运营商 & IXP',
        mtr_empty: '等待开始实时 MTR 路由追踪…',
        speedtest_title: 'HTML5 浏览器直连测速',
        speedtest_desc: '纯内存伪随机不可压缩流直连测试，零磁盘落盘开销。',
        start_speedtest: '开始测速',
        dl_mbps: '下载带宽 (Mbps)',
        ul_mbps: '上传带宽 (Mbps)',
        testfiles_title: '零落盘测速文件下载',
        testfiles_desc: '服务端内存实时流式生成，支持浏览器或命令行拉取。',
        iperf_title: 'iPerf3 临时端口租约',
        iperf_desc: '按需分配独占临时端口并启动单会话 iPerf3 服务端，到期自动销毁防滥用。',
        lease_iperf: '申请临时 iPerf3 端口',
        lease_port: '分配端口：',
        copy: '复制',
        copied: '已复制!',
        share_title: '论坛直贴 Markdown 报告导出',
        share_desc: '自动汇总最近一次多节点横评、MTR 路由及测速结果，默认对公网目标 IP 末段打码保护隐私。',
        redact_ip: '自动脱敏目标 IP',
        copy_md: '一键复制 Markdown',
        view_hops: '查看跳点',
        hide_hops: '收起',
        probing: '探测中...',
        bgp_title: 'BGP 路由表查询与 AS-Path 拓扑图',
        bgp_desc: '只读对接 BIRD 2/3、FRRouting、GoBGP，支持前缀路由优选分析、Large Community 解码与纯 SVG 自治系统路径拓扑。',
        bgp_peers_btn: '查看 BGP 邻居会话',
        bgp_disabled: '当前选中节点未开启 BGP 模块（可在配置中设置 [bgp] enabled = true 并指定 bird / frr / gobgp 驱动）。',
        bgp_raw: '查看路由器原始输出 (Raw RIB Output)',
    },
    'en-US': {
        your_ip: 'Your IP',
        probe_me: 'Probe My IP',
        nodes_title: 'Available Looking Glass Nodes',
        refresh: 'Refresh',
        tab_compare: '⚡ Multi-Node Compare',
        tab_diag: '🧭 Live MTR & Ping',
        tab_speed: '🚀 Speedtest & iPerf3',
        tab_bgp: '🌐 BGP & AS-Path',
        tab_share: '📋 Share Markdown',
        target_label: 'Target IP or Hostname',
        ip_version: 'IP Stack',
        diag_tool: 'Diagnostic Tool',
        start_test: 'Run Probe',
        stop_test: 'Stop',
        compare_heading: 'Multi-Region Latency & Return Route Comparison',
        compare_desc: 'Test your IP once across all online nodes concurrently with automatic carrier backbone & IXP classification.',
        include_route: 'Trace Return Route',
        col_node: 'Node',
        col_route: 'Return Route',
        col_avg_rtt: 'Avg Latency',
        col_minmax: 'Min / Max / Jitter',
        col_loss: 'Loss',
        col_detail: 'Hops',
        compare_empty: 'Click "Probe My IP" or enter a target above and press "Run Probe" to compare all online nodes.',
        diag_current_node: 'Active Node: ',
        diag_ready: 'Select a node above and start real-time streaming MTR or Ping.',
        stat_avg: 'Avg RTT',
        stat_min_max: 'Min / Max',
        stat_jitter: 'Jitter',
        stat_loss: 'Packet Loss',
        detected_route: 'Classified Route: ',
        col_hop_host: 'Hop IP / PTR',
        col_asn_org: 'ASN / Carrier & IXP',
        mtr_empty: 'Ready to start live streaming MTR...',
        speedtest_title: 'HTML5 Browser Speedtest',
        speedtest_desc: 'Direct zero-disk incompressible ChaCha8 stream benchmark.',
        start_speedtest: 'Start Speedtest',
        dl_mbps: 'Download (Mbps)',
        ul_mbps: 'Upload (Mbps)',
        testfiles_title: 'Zero-Disk Test Files',
        testfiles_desc: 'Generated on-the-fly in memory with zero disk footprint.',
        iperf_title: 'Ephemeral iPerf3 Port Lease',
        iperf_desc: 'Allocates an isolated temporary port running one-shot iPerf3 with auto-expiry.',
        lease_iperf: 'Request Temporary iPerf3 Port',
        lease_port: 'Assigned Port:',
        copy: 'Copy',
        copied: 'Copied!',
        share_title: 'Redacted Forum Markdown Export',
        share_desc: 'Formats your latest Compare, MTR, Ping, and Speedtest metrics into clean Markdown with IP masking.',
        redact_ip: 'Redact Target IP Octet',
        copy_md: 'Copy Markdown',
        view_hops: 'Hops',
        hide_hops: 'Hide',
        probing: 'Probing...',
        bgp_title: 'BGP RIB Lookup & AS-Path Topology',
        bgp_desc: 'Read-only BIRD 2/3, FRRouting, and GoBGP route inspector with community decoding and pure SVG AS-Path graph.',
        bgp_peers_btn: 'View BGP Peers',
        bgp_disabled: 'BGP Looking Glass is not enabled on this node (enable [bgp] in config.toml).',
        bgp_raw: 'View Raw Router RIB Output',
    }
};

const state = {
    locale: localStorage.getItem('vmlg-locale') || 'zh-CN',
    activeTab: 'compare',
    meta: null,
    nodes: [],
    selectedNodeId: '',
    activeSSE: null,
    speedAbort: null,
    iperfTimer: null,
    compareResults: {}, // nodeId -> { node, ping, mtr, error, expanded }
    lastTarget: '',
    lastPing: null,
    lastMTR: null,
    lastSpeed: null,
    lastBGP: null,
};

function t(key) {
    const dict = I18N[state.locale] || I18N['zh-CN'];
    return dict[key] || key;
}

function escapeHTML(str) {
    return String(str ?? '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
}

function redactIP(raw) {
    if (!raw) return '';
    const s = String(raw).trim();
    if (/^\d{1,3}(\.\d{1,3}){3}$/.test(s)) {
        const p = s.split('.');
        return `${p[0]}.${p[1]}.${p[2]}.x`;
    }
    if (s.includes(':')) {
        const parts = s.split(':').filter(Boolean);
        return parts.slice(0, 3).join(':') + '::x';
    }
    return s;
}

function applyLocale() {
    document.documentElement.lang = state.locale;
    document.getElementById('btn-lang').textContent = state.locale === 'zh-CN' ? 'EN' : '中文';
    document.querySelectorAll('[data-i18n]').forEach((el) => {
        el.textContent = t(el.getAttribute('data-i18n'));
    });
    renderShareMarkdown();
}

function selectedNode() {
    return state.nodes.find((n) => n.id === state.selectedNodeId) || state.nodes[0] || null;
}

function nodeSpeedBaseURL(node) {
    if (!node) return '';
    if (node.public_url) return node.public_url.replace(/\/+$/, '');
    if (node.is_local) return window.location.origin;
    return '';
}

function withTicket(url, node) {
    if (!node || !node.speedtest_ticket) return url;
    const sep = url.includes('?') ? '&' : '?';
    return `${url}${sep}ticket=${encodeURIComponent(node.speedtest_ticket)}`;
}

async function loadMeta(preserveSelection = false) {
    const res = await fetch('/api/v1/meta');
    if (!res.ok) return;
    const meta = await res.json();
    state.meta = meta;
    state.nodes = meta.nodes || [];

    if (!localStorage.getItem('vmlg-locale') && meta.default_locale) {
        state.locale = meta.default_locale;
    }

    document.title = meta.title || 'vmlg Looking Glass';
    document.getElementById('portal-title').textContent = meta.title || 'vmlg';
    document.getElementById('portal-subtitle').textContent = meta.subtitle || '';
    document.getElementById('portal-mode').textContent = (meta.mode || 'standalone').toUpperCase();
    document.getElementById('footer-version').textContent = meta.version ? `v${meta.version}` : '';

    if (meta.visitor && meta.visitor.ip) {
        document.getElementById('visitor-ip').textContent = meta.visitor.ip;
        const asnEl = document.getElementById('visitor-asn');
        const label = meta.visitor.line_tag || meta.visitor.asn || '';
        if (label) {
            asnEl.textContent = label;
            asnEl.hidden = false;
        }
        const targetInput = document.getElementById('target-input');
        if (!targetInput.value) {
            targetInput.value = meta.visitor.ip;
        }
    }

    if (!preserveSelection || !state.nodes.some((n) => n.id === state.selectedNodeId)) {
        const firstOnline = state.nodes.find((n) => n.online) || state.nodes[0];
        state.selectedNodeId = firstOnline ? firstOnline.id : '';
    }

    applyLocale();
    renderNodes();
    syncSelectedNodeUI();
}

function renderNodes() {
    const grid = document.getElementById('node-grid');
    const onlineCount = state.nodes.filter((n) => n.online).length;
    document.getElementById('nodes-online-count').textContent = `${onlineCount} / ${state.nodes.length}`;

    if (state.nodes.length === 0) {
        grid.innerHTML = `<div class="muted-text">No nodes registered yet.</div>`;
        return;
    }

    grid.innerHTML = state.nodes.map((n) => {
        const isSel = n.id === state.selectedNodeId;
        const tags = (n.tags || []).map((tg) => `<span class="tag-chip">${escapeHTML(tg)}</span>`).join('');
        const subLeft = escapeHTML(n.location || n.id);
        const subRight = escapeHTML(n.asn || n.provider || '');
        return `
            <div class="node-card ${isSel ? 'selected' : ''} ${n.online ? '' : 'offline'}" data-node-id="${escapeHTML(n.id)}">
                <div class="node-top">
                    <span class="node-name">
                        <span class="status-dot ${n.online ? '' : 'off'}"></span>
                        ${escapeHTML(n.name)}
                    </span>
                    <span class="node-region">${escapeHTML(n.region || 'DEFAULT')}</span>
                </div>
                <div class="node-sub">
                    <span>${subLeft}</span>
                    <span>${subRight}</span>
                </div>
                ${tags ? `<div class="node-tags">${tags}</div>` : ''}
            </div>
        `;
    }).join('');

    grid.querySelectorAll('.node-card').forEach((card) => {
        card.addEventListener('click', () => {
            state.selectedNodeId = card.getAttribute('data-node-id');
            renderNodes();
            syncSelectedNodeUI();
        });
    });
}

function syncSelectedNodeUI() {
    const node = selectedNode();
    const label = node ? `${node.name} (${node.location || node.id})` : '—';
    document.getElementById('diag-node-label').textContent = label;
    document.getElementById('speed-node-label').textContent = label;
    document.getElementById('bgp-node-label').textContent = label;

    // BGP capability
    const bgpOK = Boolean(node && node.online && node.capabilities && node.capabilities.bgp);
    document.getElementById('bgp-disabled-notice').hidden = bgpOK;
    document.getElementById('bgp-content-wrap').hidden = !bgpOK;
    document.getElementById('btn-bgp-summary').disabled = !bgpOK;
    document.getElementById('bgp-driver-pill').textContent = bgpOK
        ? `${(node.capabilities.bgp_driver || 'bgp').toUpperCase()} ${node.capabilities.router_asn ? '· ' + node.capabilities.router_asn : ''}`
        : 'DISABLED';

    // Test files
    const fileList = document.getElementById('testfiles-list');
    const base = nodeSpeedBaseURL(node);
    const files = (node && node.capabilities && node.capabilities.test_files) || [];
    const speedOK = Boolean(node && node.online && node.capabilities && node.capabilities.speedtest && base);

    document.getElementById('btn-start-speedtest').disabled = !speedOK;
    if (!speedOK) {
        fileList.innerHTML = `<div class="muted-text">Direct speedtest endpoint is not configured on this node.</div>`;
    } else {
        fileList.innerHTML = files.map((fname) => {
            const url = withTicket(`${base}/speedtest/files/${encodeURIComponent(fname)}`, node);
            const curlCmd = `curl -Lo /dev/null "${url}"`;
            return `
                <div class="file-item">
                    <a href="${escapeHTML(url)}" target="_blank" rel="noopener">⬇ ${escapeHTML(fname)}</a>
                    <button type="button" class="btn-ghost btn-sm" data-copy-raw="${escapeHTML(curlCmd)}">curl</button>
                </div>
            `;
        }).join('');
        fileList.querySelectorAll('[data-copy-raw]').forEach((btn) => {
            btn.addEventListener('click', () => copyText(btn.getAttribute('data-copy-raw'), btn));
        });
    }

    // iPerf3 capability
    const iperfCap = document.getElementById('iperf-cap-badge');
    const leaseBtn = document.getElementById('btn-lease-iperf');
    if (node && node.online && node.capabilities && node.capabilities.iperf3) {
        iperfCap.textContent = 'READY';
        leaseBtn.disabled = false;
    } else {
        iperfCap.textContent = (node && node.capabilities && node.capabilities.iperf_note) ? 'UNAVAILABLE' : 'OFFLINE';
        leaseBtn.disabled = true;
    }
}

function switchTab(tabName) {
    state.activeTab = tabName;
    document.querySelectorAll('.tab').forEach((b) => {
        b.classList.toggle('active', b.getAttribute('data-tab') === tabName);
    });
    ['compare', 'diag', 'speed', 'bgp', 'share'].forEach((id) => {
        document.getElementById(`panel-${id}`).hidden = id !== tabName;
    });
    const showTargetBar = tabName === 'compare' || tabName === 'diag' || tabName === 'bgp';
    document.getElementById('target-bar').hidden = !showTargetBar;
    document.getElementById('diag-tool-group').hidden = tabName !== 'diag';
    if (tabName === 'share') {
        renderShareMarkdown();
    }
}

function stopActiveProbe() {
    if (state.activeSSE) {
        state.activeSSE.close();
        state.activeSSE = null;
    }
    document.getElementById('btn-run-probe').hidden = false;
    document.getElementById('btn-stop-probe').hidden = true;
}

function setProbeRunning(running) {
    document.getElementById('btn-run-probe').hidden = running;
    document.getElementById('btn-stop-probe').hidden = !running;
    const errEl = document.getElementById('probe-error');
    if (running) errEl.hidden = true;
}

function showProbeError(msg) {
    const errEl = document.getElementById('probe-error');
    errEl.textContent = msg;
    errEl.hidden = !msg;
}

function startCompareProbe(target, ipVer) {
    stopActiveProbe();
    setProbeRunning(true);
    state.lastTarget = target;
    state.compareResults = {};

    const includeRoute = document.getElementById('compare-include-route').checked;
    const onlineNodes = state.nodes.filter((n) => n.online);
    onlineNodes.forEach((n) => {
        state.compareResults[n.id] = { node: n, ping: null, mtr: null, error: '', expanded: false };
    });
    renderCompareTable();

    const qs = new URLSearchParams({
        target,
        ip_version: ipVer,
        include_route: includeRoute ? 'true' : 'false',
    });
    const es = new EventSource(`/api/v1/probe/compare?${qs.toString()}`);
    state.activeSSE = es;

    es.addEventListener('node_ping', (e) => {
        const d = JSON.parse(e.data);
        if (state.compareResults[d.node_id]) {
            state.compareResults[d.node_id].ping = d.summary;
            renderCompareTable();
        }
    });

    es.addEventListener('node_route', (e) => {
        const d = JSON.parse(e.data);
        if (state.compareResults[d.node_id]) {
            state.compareResults[d.node_id].mtr = d.mtr;
            renderCompareTable();
        }
    });

    es.addEventListener('node_error', (e) => {
        const d = JSON.parse(e.data);
        if (state.compareResults[d.node_id]) {
            state.compareResults[d.node_id].error = d.error;
            renderCompareTable();
        }
    });

    es.addEventListener('error_msg', (e) => {
        const d = JSON.parse(e.data);
        showProbeError(d.error || 'Request rejected');
        stopActiveProbe();
    });

    es.addEventListener('complete', () => {
        stopActiveProbe();
        renderShareMarkdown();
    });

    es.onerror = () => {
        stopActiveProbe();
    };
}

function renderCompareTable() {
    const tbody = document.getElementById('compare-tbody');
    const items = Object.values(state.compareResults);
    if (items.length === 0) return;

    items.sort((a, b) => {
        const ra = a.ping && a.ping.recv > 0 ? a.ping.avg_ms : 999999;
        const rb = b.ping && b.ping.recv > 0 ? b.ping.avg_ms : 999999;
        return ra - rb;
    });

    const maxRtt = Math.max(100, ...items.map((it) => (it.ping && it.ping.avg_ms) || 0));

    let html = '';
    items.forEach((it) => {
        const n = it.node;
        let routeCell = `<span class="muted-text">${escapeHTML(t('probing'))}</span>`;
        if (it.mtr && it.mtr.route) {
            const r = it.mtr.route;
            const cls = r.tier === 'premium' ? 'premium' : (r.tier === 'standard' ? 'standard' : '');
            routeCell = `<span class="line-badge ${cls}">${escapeHTML(r.label)}</span>`;
        } else if (!document.getElementById('compare-include-route').checked) {
            routeCell = `<span class="muted-text">—</span>`;
        } else if (it.error && !it.mtr) {
            routeCell = `<span class="loss-bad">Error</span>`;
        }

        let rttCell = `<span class="muted-text">${escapeHTML(t('probing'))}</span>`;
        let minMaxCell = '—';
        let lossCell = '—';

        if (it.error && !it.ping) {
            rttCell = `<span class="loss-bad">${escapeHTML(it.error)}</span>`;
        } else if (it.ping) {
            const p = it.ping;
            if (p.recv > 0) {
                const pct = Math.min(100, Math.max(5, (p.avg_ms / maxRtt) * 100));
                const barCls = p.avg_ms < 60 ? 'rtt-good' : (p.avg_ms < 160 ? 'rtt-mid' : 'rtt-high');
                rttCell = `
                    <div class="rtt-bar-wrap">
                        <strong style="min-width:64px">${p.avg_ms.toFixed(1)} ms</strong>
                        <div class="rtt-track"><div class="rtt-fill ${barCls}" style="width:${pct}%"></div></div>
                    </div>
                `;
                minMaxCell = `${p.min_ms.toFixed(1)} / ${p.max_ms.toFixed(1)} ms (±${p.jitter_ms.toFixed(1)})`;
            } else {
                rttCell = `<span class="loss-bad">Timeout</span>`;
            }
            const lossCls = p.loss_pct === 0 ? 'loss-ok' : (p.loss_pct < 20 ? 'loss-warn' : 'loss-bad');
            lossCell = `<span class="${lossCls}">${p.loss_pct.toFixed(0)}%</span>`;
        }

        const canExpand = it.mtr && it.mtr.hops && it.mtr.hops.length > 0;
        const detailBtn = canExpand
            ? `<button type="button" class="btn-ghost btn-sm" data-toggle-hops="${escapeHTML(n.id)}">${escapeHTML(it.expanded ? t('hide_hops') : `${t('view_hops')} (${it.mtr.hops.length})`)}</button>`
            : '—';

        html += `
            <tr>
                <td>
                    <strong>${escapeHTML(n.name)}</strong>
                    <div class="muted-text" style="font-size:11.5px">${escapeHTML(n.location || '')} · ${escapeHTML(n.region || '')}</div>
                </td>
                <td>${routeCell}</td>
                <td>${rttCell}</td>
                <td style="font-family:var(--font-mono);font-size:12px">${minMaxCell}</td>
                <td style="font-family:var(--font-mono)">${lossCell}</td>
                <td>${detailBtn}</td>
            </tr>
        `;

        if (canExpand && it.expanded) {
            const hopSummary = it.mtr.hops.map((h) => {
                const tag = h.ixp || h.line_tag || h.asn || '';
                const addr = h.addr || '*';
                const rtt = h.recv > 0 ? `${h.avg_ms.toFixed(1)}ms` : '*';
                return `[${h.hop}] ${addr}${tag ? ` (${tag})` : ''} ${rtt}`;
            }).join('  →  ');
            html += `<tr class="hop-path-row"><td colspan="6">${escapeHTML(hopSummary)}</td></tr>`;
        }
    });

    tbody.innerHTML = html;
    tbody.querySelectorAll('[data-toggle-hops]').forEach((btn) => {
        btn.addEventListener('click', () => {
            const id = btn.getAttribute('data-toggle-hops');
            if (state.compareResults[id]) {
                state.compareResults[id].expanded = !state.compareResults[id].expanded;
                renderCompareTable();
            }
        });
    });
}

function startSingleDiag(target, ipVer) {
    const node = selectedNode();
    if (!node) return;
    stopActiveProbe();
    setProbeRunning(true);
    state.lastTarget = target;

    const tool = document.getElementById('diag-tool-select').value;
    const pingView = document.getElementById('ping-view');
    const mtrView = document.getElementById('mtr-view');
    const engPill = document.getElementById('diag-engine-pill');
    engPill.hidden = true;

    if (tool === 'mtr') {
        pingView.hidden = true;
        mtrView.hidden = false;
        document.getElementById('mtr-tbody').innerHTML = `<tr class="empty-row"><td colspan="10">${escapeHTML(t('probing'))}</td></tr>`;

        const qs = new URLSearchParams({ node: node.id, target, ip_version: ipVer });
        const es = new EventSource(`/api/v1/probe/mtr?${qs.toString()}`);
        state.activeSSE = es;

        const onSnap = (e) => {
            const snap = JSON.parse(e.data);
            state.lastMTR = { node, snap };
            renderMTRSnapshot(snap);
        };
        es.addEventListener('snapshot', onSnap);
        es.addEventListener('done', (e) => {
            onSnap(e);
            stopActiveProbe();
            renderShareMarkdown();
        });
        es.addEventListener('error_msg', (e) => {
            const d = JSON.parse(e.data);
            showProbeError(d.error || 'MTR failed');
            stopActiveProbe();
        });
        es.onerror = () => stopActiveProbe();
    } else {
        pingView.hidden = false;
        mtrView.hidden = true;
        document.getElementById('ping-bars').innerHTML = '';
        document.getElementById('ping-log').innerHTML = '';
        ['ping-avg', 'ping-minmax', 'ping-jitter', 'ping-loss'].forEach((id) => {
            document.getElementById(id).textContent = '…';
        });

        const samples = [];
        const qs = new URLSearchParams({
            node: node.id,
            target,
            protocol: tool,
            ip_version: ipVer,
        });
        const es = new EventSource(`/api/v1/probe/ping?${qs.toString()}`);
        state.activeSSE = es;

        es.addEventListener('sample', (e) => {
            const s = JSON.parse(e.data);
            samples.push(s);
            renderPingLiveSamples(samples);
        });
        es.addEventListener('summary', (e) => {
            const sum = JSON.parse(e.data);
            state.lastPing = { node, sum };
            renderPingSummary(sum);
            stopActiveProbe();
            renderShareMarkdown();
        });
        es.addEventListener('error_msg', (e) => {
            const d = JSON.parse(e.data);
            showProbeError(d.error || 'Ping failed');
            stopActiveProbe();
        });
        es.onerror = () => stopActiveProbe();
    }
}

function renderPingLiveSamples(samples) {
    const barsEl = document.getElementById('ping-bars');
    const logEl = document.getElementById('ping-log');
    const maxV = Math.max(50, ...samples.map((s) => s.rtt_ms || 0));

    barsEl.innerHTML = samples.map((s) => {
        if (s.timeout) {
            return `<div class="ping-bar-item rtt-high" style="height:100%;opacity:0.45" title="Seq ${s.seq}: Timeout"></div>`;
        }
        const h = Math.max(12, Math.min(100, (s.rtt_ms / maxV) * 100));
        const cls = s.rtt_ms < 60 ? 'rtt-good' : (s.rtt_ms < 160 ? 'rtt-mid' : 'rtt-high');
        return `<div class="ping-bar-item ${cls}" style="height:${h}%" title="Seq ${s.seq}: ${s.rtt_ms} ms"></div>`;
    }).join('');

    const latest = samples[samples.length - 1];
    const line = document.createElement('div');
    if (latest.timeout) {
        line.className = 'loss-bad';
        line.textContent = `[seq=${latest.seq}] ${latest.addr} (${latest.proto}) Request timeout (${latest.error || 'timeout'})`;
    } else {
        const ttlPart = latest.ttl ? ` ttl=${latest.ttl}` : '';
        line.textContent = `[seq=${latest.seq}] Reply from ${latest.addr} (${latest.proto}):${ttlPart} time=${latest.rtt_ms.toFixed(2)} ms`;
    }
    logEl.appendChild(line);
    logEl.scrollTop = logEl.scrollHeight;
}

function renderPingSummary(sum) {
    const engPill = document.getElementById('diag-engine-pill');
    engPill.textContent = sum.engine;
    engPill.hidden = false;

    document.getElementById('ping-avg').textContent = sum.recv > 0 ? `${sum.avg_ms.toFixed(2)} ms` : 'Timeout';
    document.getElementById('ping-minmax').textContent = sum.recv > 0 ? `${sum.min_ms.toFixed(2)} / ${sum.max_ms.toFixed(2)} ms` : '—';
    document.getElementById('ping-jitter').textContent = sum.recv > 0 ? `${sum.jitter_ms.toFixed(2)} ms` : '—';
    document.getElementById('ping-loss').textContent = `${sum.loss_pct.toFixed(0)}% (${sum.sent - sum.recv}/${sum.sent})`;
}

function renderMTRSnapshot(snap) {
    const engPill = document.getElementById('diag-engine-pill');
    engPill.textContent = snap.engine;
    engPill.hidden = false;

    const banner = document.getElementById('mtr-route-banner');
    banner.hidden = false;
    const badge = document.getElementById('mtr-verdict-badge');
    badge.textContent = snap.route ? snap.route.label : 'BGP / Direct';
    badge.className = `line-badge ${snap.route && snap.route.tier === 'premium' ? 'premium' : ''}`;
    document.getElementById('mtr-cycle-progress').textContent = `Cycle ${snap.cycle} / ${snap.max_cycles}${snap.completed ? ' ✓' : ''}`;

    const tbody = document.getElementById('mtr-tbody');
    tbody.innerHTML = (snap.hops || []).map((h) => {
        const hostLine = h.addr
            ? `<div><strong>${escapeHTML(h.addr)}</strong></div>${h.ptr ? `<div class="muted-text" style="font-size:11px">${escapeHTML(h.ptr)}</div>` : ''}`
            : `<span class="muted-text">* (No reply)</span>`;

        let asnBadges = '';
        if (h.ixp) {
            asnBadges += `<span class="line-badge ixp">${escapeHTML(h.ixp)}</span> `;
        } else if (h.line_tag) {
            const cls = h.tier === 'premium' ? 'premium' : (h.tier === 'standard' ? 'standard' : '');
            asnBadges += `<span class="line-badge ${cls}">${escapeHTML(h.line_tag)}</span> `;
        }
        const orgText = [h.asn, h.org].filter(Boolean).join(' ');
        if (orgText) {
            asnBadges += `<span class="muted-text" style="font-size:12px">${escapeHTML(orgText)}</span>`;
        }

        const lossCls = h.loss_pct === 0 ? 'loss-ok' : (h.loss_pct < 30 ? 'loss-warn' : 'loss-bad');
        const fmt = (v) => (h.recv > 0 ? v.toFixed(1) : '—');

        return `
            <tr>
                <td>${h.hop}</td>
                <td style="font-family:var(--font-mono)">${hostLine}</td>
                <td style="font-family:var(--font-sans)">${asnBadges || '—'}</td>
                <td class="${lossCls}">${h.loss_pct.toFixed(0)}%</td>
                <td>${h.sent}</td>
                <td>${fmt(h.last_ms)}</td>
                <td><strong>${fmt(h.avg_ms)}</strong></td>
                <td>${fmt(h.best_ms)}</td>
                <td>${fmt(h.wrst_ms)}</td>
                <td>${fmt(h.stdev_ms)}</td>
            </tr>
        `;
    }).join('');
}

// HTML5 Browser Speedtest
async function runBrowserSpeedtest() {
    const node = selectedNode();
    const base = nodeSpeedBaseURL(node);
    if (!node || !base) return;

    const ac = new AbortController();
    state.speedAbort = ac;
    document.getElementById('btn-start-speedtest').hidden = true;
    document.getElementById('btn-stop-speedtest').hidden = false;
    ['st-dl', 'st-ul', 'st-ping', 'st-jitter'].forEach((id) => {
        document.getElementById(id).textContent = '…';
    });

    const dlPoints = [];
    const ulPoints = [];
    drawSpeedChart(dlPoints, ulPoints);

    const setPhase = (txt) => {
        document.getElementById('st-phase-text').textContent = txt;
    };

    try {
        // Phase 1: Ping & Jitter (8 samples)
        setPhase('1/3 Measuring HTTP Latency & Jitter...');
        const rtts = [];
        for (let i = 0; i < 8; i++) {
            const url = withTicket(`${base}/api/v1/speedtest/ping?t=${Date.now()}-${i}`, node);
            const t0 = performance.now();
            await fetch(url, { cache: 'no-store', signal: ac.signal });
            const rtt = performance.now() - t0;
            if (i > 0) rtts.push(rtt); // discard cold connection handshake sample 0
        }
        const avgPing = rtts.reduce((a, b) => a + b, 0) / rtts.length;
        let jitter = 0;
        for (let i = 1; i < rtts.length; i++) jitter += Math.abs(rtts[i] - rtts[i - 1]);
        jitter = rtts.length > 1 ? jitter / (rtts.length - 1) : 0;

        document.getElementById('st-ping').textContent = avgPing.toFixed(1);
        document.getElementById('st-jitter').textContent = jitter.toFixed(1);

        // Phase 2: Download (multi-stream 3 parallel streams, ~5 seconds)
        setPhase('2/3 Measuring Download Throughput...');
        const dlMbps = await measureDownloadMbps(base, node, ac.signal, (mbps) => {
            dlPoints.push(mbps);
            document.getElementById('st-dl').textContent = mbps.toFixed(1);
            drawSpeedChart(dlPoints, ulPoints);
        });
        document.getElementById('st-dl').textContent = dlMbps.toFixed(1);

        // Phase 3: Upload (~4 seconds)
        setPhase('3/3 Measuring Upload Throughput...');
        const ulMbps = await measureUploadMbps(base, node, ac.signal, (mbps) => {
            ulPoints.push(mbps);
            document.getElementById('st-ul').textContent = mbps.toFixed(1);
            drawSpeedChart(dlPoints, ulPoints);
        });
        document.getElementById('st-ul').textContent = ulMbps.toFixed(1);

        state.lastSpeed = {
            node,
            dlMbps: dlMbps.toFixed(1),
            ulMbps: ulMbps.toFixed(1),
            pingMs: avgPing.toFixed(1),
            jitterMs: jitter.toFixed(1),
        };
        setPhase('✓ Speedtest completed.');
        renderShareMarkdown();
    } catch (err) {
        if (err.name === 'AbortError') {
            setPhase('Speedtest stopped.');
        } else {
            setPhase(`Speedtest error: ${err.message}`);
        }
    } finally {
        state.speedAbort = null;
        document.getElementById('btn-start-speedtest').hidden = false;
        document.getElementById('btn-stop-speedtest').hidden = true;
    }
}

async function measureDownloadMbps(base, node, outerSignal, onTick) {
    const durationMs = 4500;
    const localAC = new AbortController();
    const onAbort = () => localAC.abort();
    outerSignal.addEventListener('abort', onAbort, { once: true });

    let totalBytes = 0;
    const t0 = performance.now();
    const timer = setInterval(() => {
        const sec = (performance.now() - t0) / 1000;
        if (sec > 0.15) onTick((totalBytes * 8) / (sec * 1e6));
        if (performance.now() - t0 >= durationMs) localAC.abort();
    }, 200);

    const worker = async (idx) => {
        while (performance.now() - t0 < durationMs && !localAC.signal.aborted) {
            const url = withTicket(`${base}/api/v1/speedtest/download?bytes=33554432&s=${idx}-${Date.now()}`, node);
            try {
                const resp = await fetch(url, { cache: 'no-store', signal: localAC.signal });
                if (!resp.body) break;
                const reader = resp.body.getReader();
                while (true) {
                    const { done, value } = await reader.read();
                    if (done) break;
                    totalBytes += value.byteLength;
                }
            } catch (e) {
                if (localAC.signal.aborted && !outerSignal.aborted) break;
                if (outerSignal.aborted) throw e;
                break;
            }
        }
    };

    try {
        await Promise.all([worker(0), worker(1), worker(2)]);
    } finally {
        clearInterval(timer);
        outerSignal.removeEventListener('abort', onAbort);
    }
    if (outerSignal.aborted) throw new DOMException('Aborted', 'AbortError');
    const elapsedSec = Math.max(0.05, (performance.now() - t0) / 1000);
    return (totalBytes * 8) / (elapsedSec * 1e6);
}

async function measureUploadMbps(base, node, signal, onTick) {
    const chunk = new Uint8Array(2 * 1024 * 1024); // 2 MiB pseudo-random buffer
    for (let i = 0; i < chunk.length; i += 4096) {
        chunk[i] = (i * 31) & 0xff;
    }
    const durationMs = 4000;
    let totalBytes = 0;
    const t0 = performance.now();

    while (performance.now() - t0 < durationMs) {
        if (signal.aborted) throw new DOMException('Aborted', 'AbortError');
        const url = withTicket(`${base}/api/v1/speedtest/upload?t=${Date.now()}`, node);
        const resp = await fetch(url, {
            method: 'POST',
            body: chunk,
            signal,
            headers: { 'Content-Type': 'application/octet-stream' },
        });
        if (!resp.ok) break;
        totalBytes += chunk.byteLength;
        const sec = (performance.now() - t0) / 1000;
        if (sec > 0.1) onTick((totalBytes * 8) / (sec * 1e6));
    }
    const elapsedSec = Math.max(0.05, (performance.now() - t0) / 1000);
    return (totalBytes * 8) / (elapsedSec * 1e6);
}

function drawSpeedChart(dl, ul) {
    const canvas = document.getElementById('speed-canvas');
    const ctx = canvas.getContext('2d');
    const w = canvas.width;
    const h = canvas.height;
    ctx.clearRect(0, 0, w, h);

    const maxVal = Math.max(50, ...dl, ...ul) * 1.15;
    ctx.strokeStyle = 'rgba(139, 155, 180, 0.18)';
    ctx.lineWidth = 1;
    for (let i = 1; i <= 3; i++) {
        const y = (h / 4) * i;
        ctx.beginPath();
        ctx.moveTo(0, y);
        ctx.lineTo(w, y);
        ctx.stroke();
    }

    const plotLine = (pts, color, xOffsetFrac, widthFrac) => {
        if (pts.length < 2) return;
        ctx.strokeStyle = color;
        ctx.lineWidth = 2.5;
        ctx.beginPath();
        pts.forEach((v, idx) => {
            const x = w * xOffsetFrac + (idx / Math.max(1, pts.length - 1)) * (w * widthFrac);
            const y = h - 10 - (v / maxVal) * (h - 24);
            if (idx === 0) ctx.moveTo(x, y);
            else ctx.lineTo(x, y);
        });
        ctx.stroke();
    };

    plotLine(dl, '#38bdf8', 0.02, 0.46);
    plotLine(ul, '#10b981', 0.52, 0.46);
}

// iPerf3 Lease
async function leaseIPerf3() {
    const node = selectedNode();
    if (!node) return;
    const errEl = document.getElementById('iperf-error');
    errEl.hidden = true;

    const btn = document.getElementById('btn-lease-iperf');
    btn.disabled = true;
    try {
        const res = await fetch(`/api/v1/iperf3/lease?node=${encodeURIComponent(node.id)}`, { method: 'POST' });
        const data = await res.json();
        if (!res.ok) {
            errEl.textContent = data.error || 'Failed to acquire iPerf3 port';
            errEl.hidden = false;
            return;
        }

        document.getElementById('iperf-lease-box').hidden = false;
        document.getElementById('iperf-port-val').textContent = `${data.host}:${data.port}`;
        document.getElementById('iperf-cmd-tcp').textContent = data.command_tcp;
        document.getElementById('iperf-cmd-rev').textContent = data.command_reverse;

        if (state.iperfTimer) clearInterval(state.iperfTimer);
        const expMs = new Date(data.expires_at).getTime();
        const tick = () => {
            const rem = Math.max(0, Math.round((expMs - Date.now()) / 1000));
            document.getElementById('iperf-countdown').textContent = `${rem}s`;
            if (rem <= 0) {
                clearInterval(state.iperfTimer);
                document.getElementById('iperf-lease-box').hidden = true;
            }
        };
        tick();
        state.iperfTimer = setInterval(tick, 1000);
    } catch (e) {
        errEl.textContent = e.message;
        errEl.hidden = false;
    } finally {
        btn.disabled = false;
    }
}

// BGP Route Lookup & Pure SVG AS-Path Topology
async function runBGPRouteLookup(target) {
    const node = selectedNode();
    if (!node) return;
    showProbeError('');
    setProbeRunning(true);
    state.lastTarget = target;

    try {
        const qs = new URLSearchParams({ node: node.id, target });
        const res = await fetch(`/api/v1/bgp/route?${qs.toString()}`);
        const data = await res.json();
        if (!res.ok) {
            showProbeError(data.error || 'BGP lookup failed');
            return;
        }
        state.lastBGP = { node, res: data };
        renderBGPRouteResult(data, node);
        renderShareMarkdown();
    } catch (e) {
        showProbeError(e.message);
    } finally {
        setProbeRunning(false);
    }
}

async function loadBGPSummary() {
    const node = selectedNode();
    if (!node) return;
    showProbeError('');
    const btn = document.getElementById('btn-bgp-summary');
    btn.disabled = true;
    try {
        const res = await fetch(`/api/v1/bgp/summary?node=${encodeURIComponent(node.id)}`);
        const data = await res.json();
        if (!res.ok) {
            showProbeError(data.error || 'BGP summary failed');
            return;
        }
        renderBGPPeers(data);
    } catch (e) {
        showProbeError(e.message);
    } finally {
        btn.disabled = false;
    }
}

function renderBGPRouteResult(data, node) {
    const paths = data.paths || [];
    const svgWrap = document.getElementById('bgp-svg-wrap');
    const pathsWrap = document.getElementById('bgp-paths-wrap');
    const tbody = document.getElementById('bgp-paths-tbody');
    const rawDetails = document.getElementById('bgp-raw-details');

    if (paths.length > 0) {
        svgWrap.innerHTML = buildASPathSVG(data.router_asn || node.asn || node.name, data.prefix || data.target, paths);
        svgWrap.hidden = false;
        pathsWrap.hidden = false;

        tbody.innerHTML = paths.map((p) => {
            const bestCell = p.best
                ? `<span class="line-badge premium">★ BEST</span>`
                : `<span class="line-badge subtle">ALT</span>`;
            const asChips = (p.as_path || []).map((hop) => {
                const tip = hop.name ? ` title="${escapeHTML(hop.name)}"` : '';
                return `<span class="tag-chip"${tip}>${escapeHTML(hop.asn)}${hop.name ? ' · ' + escapeHTML(hop.name) : ''}</span>`;
            }).join(' → ') || 'Local / IGP';

            const commChips = (p.communities || []).map((c) => {
                const txt = c.label ? `${c.raw} (${c.label})` : c.raw;
                return `<span class="line-badge subtle" style="margin:2px">${escapeHTML(txt)}</span>`;
            }).join(' ');

            return `
                <tr>
                    <td>${bestCell}</td>
                    <td style="font-family:var(--font-mono)">${escapeHTML(p.protocol || 'BGP')}</td>
                    <td style="font-family:var(--font-mono)">${escapeHTML(p.next_hop || '—')}${p.interface ? ` <span class="muted-text">(${escapeHTML(p.interface)})</span>` : ''}</td>
                    <td style="font-family:var(--font-mono)">${p.local_pref || '—'} / ${p.med ?? 0}</td>
                    <td>${asChips}</td>
                    <td>${commChips || '—'}</td>
                </tr>
            `;
        }).join('');
    } else {
        svgWrap.hidden = true;
        pathsWrap.hidden = true;
    }

    if (data.raw_output) {
        document.getElementById('bgp-raw-pre').textContent = data.raw_output;
        rawDetails.hidden = false;
    } else {
        rawDetails.hidden = true;
    }
}

function buildASPathSVG(localASN, prefix, paths) {
    // Build layered DAG columns: col 0 = localASN, intermediate hops, last col = prefix/origin
    const maxHops = Math.max(1, ...paths.map((p) => (p.as_path || []).length));
    const totalCols = maxHops + 1;

    const nodesByKey = new Map();
    const edges = [];

    const ensureNode = (id, title, sub, col, isBest) => {
        const key = `${col}:${id}`;
        if (!nodesByKey.has(key)) {
            nodesByKey.set(key, { key, id, title, sub, col, best: isBest });
        } else if (isBest) {
            nodesByKey.get(key).best = true;
        }
        return key;
    };

    const rootKey = ensureNode(localASN, localASN, 'Looking Glass Router', 0, true);

    paths.forEach((p) => {
        const hops = p.as_path || [];
        let prevKey = rootKey;
        hops.forEach((h, idx) => {
            const isLast = idx === hops.length - 1;
            const col = isLast ? maxHops : idx + 1;
            const sub = isLast ? `${h.name || 'Origin'} (${prefix})` : (h.name || 'Transit');
            const curKey = ensureNode(h.asn, h.asn, sub, col, Boolean(p.best));
            edges.push({ from: prevKey, to: curKey, best: Boolean(p.best) });
            prevKey = curKey;
        });
    });

    // Group by column to assign vertical Y coordinates
    const cols = Array.from({ length: totalCols }, () => []);
    nodesByKey.forEach((n) => cols[n.col].push(n));

    const maxRows = Math.max(1, ...cols.map((c) => c.length));
    const boxW = 158;
    const boxH = 48;
    const colGap = 210;
    const rowGap = 72;
    const width = Math.max(620, 48 + (totalCols - 1) * colGap + boxW);
    const height = Math.max(120, 36 + maxRows * rowGap);

    cols.forEach((colNodes, cIdx) => {
        const totalH = (colNodes.length - 1) * rowGap;
        const startY = (height - boxH) / 2 - totalH / 2;
        colNodes.forEach((n, rIdx) => {
            n.x = 24 + cIdx * colGap;
            n.y = startY + rIdx * rowGap;
        });
    });

    const edgeSVG = edges.map((e) => {
        const a = nodesByKey.get(e.from);
        const b = nodesByKey.get(e.to);
        const x1 = a.x + boxW;
        const y1 = a.y + boxH / 2;
        const x2 = b.x;
        const y2 = b.y + boxH / 2;
        const midX = (x1 + x2) / 2;
        const stroke = e.best ? '#10b981' : '#64748b';
        const dash = e.best ? '' : 'stroke-dasharray="5,4"';
        const sw = e.best ? 2.6 : 1.5;
        return `<path d="M ${x1} ${y1} C ${midX} ${y1}, ${midX} ${y2}, ${x2} ${y2}" fill="none" stroke="${stroke}" stroke-width="${sw}" ${dash}/>`;
    }).join('');

    const nodeSVG = Array.from(nodesByKey.values()).map((n) => {
        const border = n.best ? '#10b981' : '#334155';
        const subTrunc = n.sub.length > 22 ? n.sub.slice(0, 21) + '…' : n.sub;
        return `
            <g transform="translate(${n.x}, ${n.y})">
                <rect width="${boxW}" height="${boxH}" rx="8" fill="#151d2e" stroke="${border}" stroke-width="1.6"/>
                <text x="12" y="21" fill="#e6edf7" font-family="monospace" font-size="12.5" font-weight="700">${escapeHTML(n.title)}</text>
                <text x="12" y="37" fill="#8b9bb4" font-size="10.5">${escapeHTML(subTrunc)}</text>
            </g>
        `;
    }).join('');

    return `<svg viewBox="0 0 ${width} ${height}" xmlns="http://www.w3.org/2000/svg">${edgeSVG}${nodeSVG}</svg>`;
}

function renderBGPPeers(data) {
    const wrap = document.getElementById('bgp-peers-wrap');
    const tbody = document.getElementById('bgp-peers-tbody');
    const peers = data.peers || [];
    wrap.hidden = false;

    if (peers.length === 0) {
        tbody.innerHTML = `<tr class="empty-row"><td colspan="6">No BGP peer sessions returned.</td></tr>`;
        return;
    }
    tbody.innerHTML = peers.map((p) => {
        const up = /estab|up/i.test(p.state || '');
        const stateBadge = `<span class="line-badge ${up ? 'premium' : 'subtle'}">${escapeHTML(p.state || 'Unknown')}</span>`;
        const asnOrg = [p.asn, p.org].filter(Boolean).join(' · ') || p.info || '—';
        return `
            <tr>
                <td><strong>${escapeHTML(p.name)}</strong></td>
                <td>${escapeHTML(p.neighbor || '—')}</td>
                <td style="font-family:var(--font-sans)">${escapeHTML(asnOrg)}</td>
                <td>${stateBadge}</td>
                <td>${escapeHTML(p.since || '—')}</td>
                <td>${p.prefixes_rcvd ? p.prefixes_rcvd.toLocaleString() : '—'}</td>
            </tr>
        `;
    }).join('');
}

// Share Markdown Builder
function renderShareMarkdown() {
    const redact = document.getElementById('share-redact-toggle').checked;
    const fmtTarget = (raw) => (redact ? redactIP(raw) : raw);
    const lines = [];
    const title = (state.meta && state.meta.title) || 'vmlg Looking Glass';
    lines.push(`### ${title} - Network Diagnostic Report`);
    lines.push(`- Generated: \`${new Date().toISOString()}\``);
    if (state.lastTarget) {
        lines.push(`- Target: \`${fmtTarget(state.lastTarget)}\``);
    }
    lines.push('');

    const cmpItems = Object.values(state.compareResults).filter((x) => x.ping || x.mtr);
    if (cmpItems.length > 0) {
        lines.push('#### Multi-Node Latency & Return Route');
        lines.push('| Node | Region | Return Route | Avg RTT | Min / Max | Loss |');
        lines.push('| :--- | :--- | :--- | ---: | ---: | ---: |');
        cmpItems.forEach((it) => {
            const route = (it.mtr && it.mtr.route && it.mtr.route.label) || '—';
            const avg = it.ping && it.ping.recv > 0 ? `${it.ping.avg_ms.toFixed(1)} ms` : 'Timeout';
            const mm = it.ping && it.ping.recv > 0 ? `${it.ping.min_ms.toFixed(1)} / ${it.ping.max_ms.toFixed(1)} ms` : '—';
            const loss = it.ping ? `${it.ping.loss_pct.toFixed(0)}%` : '—';
            lines.push(`| ${it.node.name} | ${it.node.region || '—'} | **${route}** | ${avg} | ${mm} | ${loss} |`);
        });
        lines.push('');
    }

    if (state.lastMTR && state.lastMTR.snap) {
        const { node, snap } = state.lastMTR;
        lines.push(`#### MTR Trace from ${node.name} (${snap.route ? snap.route.label : 'Direct'})`);
        lines.push('| Hop | IP | Carrier / ASN | Loss% | Avg (ms) | Best / Wrst |');
        lines.push('| ---: | :--- | :--- | ---: | ---: | ---: |');
        const hops = snap.hops || [];
        hops.forEach((h, idx) => {
            const isLast = idx === hops.length - 1;
            const ipDisp = h.addr ? (isLast && redact ? redactIP(h.addr) : h.addr) : '*';
            const tag = [h.ixp || h.line_tag, h.asn].filter(Boolean).join(' ') || '—';
            const avg = h.recv > 0 ? h.avg_ms.toFixed(1) : '—';
            const bw = h.recv > 0 ? `${h.best_ms.toFixed(1)} / ${h.wrst_ms.toFixed(1)}` : '—';
            lines.push(`| ${h.hop} | \`${ipDisp}\` | ${tag} | ${h.loss_pct.toFixed(0)}% | ${avg} | ${bw} |`);
        });
        lines.push('');
    }

    if (state.lastPing && state.lastPing.sum) {
        const { node, sum } = state.lastPing;
        lines.push(`#### Ping from ${node.name} (${sum.protocol})`);
        lines.push(`- Sent/Recv: \`${sum.sent}/${sum.recv}\` (Loss \`${sum.loss_pct}%\`) · Avg \`${sum.avg_ms} ms\` · Jitter \`${sum.jitter_ms} ms\``);
        lines.push('');
    }

    if (state.lastSpeed) {
        const s = state.lastSpeed;
        lines.push(`#### HTML5 Speedtest (${s.node.name})`);
        lines.push(`- Download: **${s.dlMbps} Mbps** · Upload: **${s.ulMbps} Mbps** · Ping: \`${s.pingMs} ms\` (Jitter \`${s.jitterMs} ms\`)`);
        lines.push('');
    }

    if (state.lastBGP && state.lastBGP.res) {
        const { node, res } = state.lastBGP;
        lines.push(`#### BGP RIB Lookup on ${node.name} (\`${fmtTarget(res.prefix || res.target)}\`)`);
        (res.paths || []).forEach((p) => {
            const mark = p.best ? '★ BEST' : 'BACKUP';
            const asChain = [res.router_asn || 'Local', ...(p.as_path || []).map((a) => a.asn)].join(' → ');
            lines.push(`- **${mark}** via \`${p.next_hop || '—'}\` (LocalPref \`${p.local_pref || '—'}\`): \`${asChain}\``);
        });
        lines.push('');
    }

    if (cmpItems.length === 0 && !state.lastMTR && !state.lastPing && !state.lastSpeed && !state.lastBGP) {
        lines.push('_Run a Multi-Node Compare, MTR, Ping, or Speedtest first to populate this report._');
    }

    document.getElementById('share-markdown-box').value = lines.join('\n');
}

async function copyText(text, btn) {
    try {
        await navigator.clipboard.writeText(text);
        if (btn) {
            const prev = btn.textContent;
            btn.textContent = t('copied');
            setTimeout(() => { btn.textContent = prev; }, 1400);
        }
    } catch (_) {}
}

function bindEvents() {
    document.querySelectorAll('.tab').forEach((btn) => {
        btn.addEventListener('click', () => switchTab(btn.getAttribute('data-tab')));
    });

    document.getElementById('btn-lang').addEventListener('click', () => {
        state.locale = state.locale === 'zh-CN' ? 'en-US' : 'zh-CN';
        localStorage.setItem('vmlg-locale', state.locale);
        applyLocale();
        renderCompareTable();
    });

    document.getElementById('btn-theme').addEventListener('click', () => {
        const cur = document.documentElement.getAttribute('data-theme') || 'dark';
        const next = cur === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', next);
        localStorage.setItem('vmlg-theme', next);
    });

    document.getElementById('btn-use-my-ip').addEventListener('click', () => {
        if (state.meta && state.meta.visitor && state.meta.visitor.ip) {
            document.getElementById('target-input').value = state.meta.visitor.ip;
            document.getElementById('probe-form').requestSubmit();
        }
    });

    document.querySelectorAll('.chip[data-fill]').forEach((chip) => {
        chip.addEventListener('click', () => {
            document.getElementById('target-input').value = chip.getAttribute('data-fill');
        });
    });

    document.getElementById('btn-refresh-nodes').addEventListener('click', () => loadMeta(true));

    document.getElementById('probe-form').addEventListener('submit', (e) => {
        e.preventDefault();
        const target = document.getElementById('target-input').value.trim();
        const ipVer = document.getElementById('ipver-select').value;
        if (!target) return;
        if (state.activeTab === 'compare') {
            startCompareProbe(target, ipVer);
        } else if (state.activeTab === 'bgp') {
            runBGPRouteLookup(target);
        } else {
            startSingleDiag(target, ipVer);
        }
    });

    document.getElementById('btn-bgp-summary').addEventListener('click', loadBGPSummary);

    document.getElementById('btn-stop-probe').addEventListener('click', stopActiveProbe);

    document.getElementById('btn-start-speedtest').addEventListener('click', runBrowserSpeedtest);
    document.getElementById('btn-stop-speedtest').addEventListener('click', () => {
        if (state.speedAbort) state.speedAbort.abort();
    });

    document.getElementById('btn-lease-iperf').addEventListener('click', leaseIPerf3);

    document.querySelectorAll('[data-copy-target]').forEach((btn) => {
        btn.addEventListener('click', () => {
            const targetId = btn.getAttribute('data-copy-target');
            const code = document.getElementById(targetId).textContent;
            copyText(code, btn);
        });
    });

    document.getElementById('share-redact-toggle').addEventListener('change', renderShareMarkdown);
    document.getElementById('btn-copy-markdown').addEventListener('click', (e) => {
        copyText(document.getElementById('share-markdown-box').value, e.currentTarget);
    });
}

bindEvents();
loadMeta(false);
