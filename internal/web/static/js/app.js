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
        tab_dig: '🔎 DNS 解析 (Dig)',
        target_label: '目标 IP 或域名',
        ip_version: '协议栈',
        diag_tool: '诊断工具',
        dig_rrtype: '记录类型',
        dig_resolver: 'DNS 服务器',
        dig_title: 'DNS 解析查询 (dig)',
        dig_desc: '从节点视角发起完整 dig 语义查询：返回每条记录 TTL、CNAME 链与完整应答区，可指定公共 DNS 对比解析差异（留空使用节点系统解析）。',
        dig_empty: '输入域名（PTR 可直接填 IP）后点击「开始测试」发起解析…',
        dig_col_name: '记录名',
        dig_col_value: '记录值',
        tab_http: '🔗 HTTP 探测',
        http_title: 'HTTP/HTTPS 可达性探测',
        http_desc: '从节点发起受控 GET：状态码、TTFB、重定向链、TLS 证书摘要与完整响应头。响应 body 仅计入 1MiB 计时上限，绝不展示。',
        http_empty: '在上方输入完整 URL（如 https://example.com）后点击「开始测试」。若省略协议将默认 https://。',
        http_status: '状态码',
        http_total: '总耗时 (ms)',
        http_bytes: '已读取 Body',
        http_hop_url: '跳转 URL',
        http_col_header: '响应头',
        http_col_value: '值',
        http_tls: 'TLS 证书',
        http_truncated: '已截断（超 1MiB）',
        rpki_title: 'RPKI ROA 验证',
        rpki_no_roa: '该前缀+Origin 无覆盖 ROA（NotFound）',
        tab_unlock: '📺 流媒体解锁',
        unlock_title: '流媒体解锁检测',
        unlock_desc: '从本节点出口探测固定五项服务可用性（公开无登录端点，纯 Go）。结果按节点缓存。',
        unlock_empty: '点击「立即检测」从本节点探测 YouTube / Netflix / Disney+ / ChatGPT / TikTok。',
        unlock_run: '立即检测',
        unlock_cache_age: '缓存于',
        unlock_service: '服务',
        unlock_status: '状态',
        unlock_region: '区域',
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
        tab_dig: '🔎 DNS Dig',
        target_label: 'Target IP or Hostname',
        ip_version: 'IP Stack',
        diag_tool: 'Diagnostic Tool',
        dig_rrtype: 'Record Type',
        dig_resolver: 'Resolver',
        dig_title: 'DNS Query (dig)',
        dig_desc: 'Full dig semantics from the node: per-record TTLs, CNAME chains and the complete answer section. Leave the resolver empty to use the node system resolver, or set a public one (e.g. 1.1.1.1) to compare.',
        dig_empty: 'Enter a domain (or an IP for reverse PTR) and press "Run Probe".',
        dig_col_name: 'Name',
        dig_col_value: 'Value',
        tab_http: '🔗 HTTP Probe',
        http_title: 'HTTP/HTTPS Reachability Probe',
        http_desc: 'Controlled GET from the node: status code, TTFB, redirect chain, TLS certificate summary, and full response headers. Body is capped at 1 MiB for timing and never displayed.',
        http_empty: 'Enter a full URL (e.g. https://example.com) above and press "Run Probe". A missing scheme defaults to https://.',
        http_status: 'Status',
        http_total: 'Total (ms)',
        http_bytes: 'Body Read',
        http_hop_url: 'Hop URL',
        http_col_header: 'Header',
        http_col_value: 'Value',
        http_tls: 'TLS Certificate',
        http_truncated: 'Truncated (>1MiB)',
        rpki_title: 'RPKI ROA Validation',
        rpki_no_roa: 'No covering ROA for this prefix + origin (NotFound)',
        tab_unlock: '📺 Unlock Check',
        unlock_title: 'Streaming Unlock Check',
        unlock_desc: "Fixed five-service availability probe from this node's egress (public no-login endpoints, pure Go). Results are cached per node.",
        unlock_empty: 'Press "Run Check" to probe YouTube / Netflix / Disney+ / ChatGPT / TikTok from this node.',
        unlock_run: 'Run Check',
        unlock_cache_age: 'cached',
        unlock_service: 'Service',
        unlock_status: 'Status',
        unlock_region: 'Region',
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
    },
    'zh-TW': {
        your_ip: '你的 IP',
        probe_me: '測我本機',
        nodes_title: '可用測試節點',
        refresh: '重新整理',
        tab_compare: '⚡ 多節點同屏橫評',
        tab_diag: '🧭 路由追蹤 & Ping',
        tab_speed: '🚀 頻寬測速 & iPerf3',
        tab_bgp: '🌐 BGP 路由 & 拓撲',
        tab_share: '📋 脫敏報告匯出',
        tab_dig: '🔎 DNS 解析 (Dig)',
        target_label: '目標 IP 或網域',
        ip_version: '協定堆疊',
        diag_tool: '診斷工具',
        dig_rrtype: '記錄類型',
        dig_resolver: 'DNS 伺服器',
        dig_title: 'DNS 解析查詢 (dig)',
        dig_desc: '從節點視角發起完整 dig 語意查詢：返回每條記錄 TTL、CNAME 鏈與完整應答區，可指定公共 DNS 對比解析差異（留空使用節點系統解析）。',
        dig_empty: '輸入網域（PTR 可直接填 IP）後點擊「開始測試」發起解析…',
        dig_col_name: '記錄名',
        dig_col_value: '記錄值',
        tab_http: '🔗 HTTP 探測',
        http_title: 'HTTP/HTTPS 可達性探測',
        http_desc: '從節點發起受控 GET：狀態碼、TTFB、重定向鏈、TLS 憑證摘要與完整回應標頭。回應 body 僅計入 1MiB 計時上限，絕不展示。',
        http_empty: '在上方輸入完整 URL（如 https://example.com）後點擊「開始測試」。若省略協定將默認 https://。',
        http_status: '狀態碼',
        http_total: '總耗時 (ms)',
        http_bytes: '已讀取 Body',
        http_hop_url: '跳轉 URL',
        http_col_header: '回應標頭',
        http_col_value: '值',
        http_tls: 'TLS 憑證',
        http_truncated: '已截斷（超 1MiB）',
        rpki_title: 'RPKI ROA 驗證',
        rpki_no_roa: '該前綴+Origin 無覆蓋 ROA（NotFound）',
        tab_unlock: '📺 串流媒體解鎖',
        unlock_title: '串流媒體解鎖檢測',
        unlock_desc: '從本節點出口探測固定五項服務可用性（公開無登入端點，純 Go）。結果按節點快取。',
        unlock_empty: '點擊「立即檢測」從本節點探測 YouTube / Netflix / Disney+ / ChatGPT / TikTok。',
        unlock_run: '立即檢測',
        unlock_cache_age: '快取於',
        unlock_service: '服務',
        unlock_status: '狀態',
        unlock_region: '區域',
        start_test: '開始測試',
        stop_test: '停止',
        compare_heading: '全球節點並發延遲與回程線路對比',
        compare_desc: '一次輸入訪客 IP，所有在線節點同步探測延遲、掉包率並自動識別回程骨幹網類型。',
        include_route: '同步追蹤回程線路',
        col_node: '節點',
        col_route: '回程線路判定',
        col_avg_rtt: '平均延遲',
        col_minmax: '最值 / 抖動',
        col_loss: '掉包率',
        col_detail: '路徑詳情',
        compare_empty: '點擊上方「測我本機」或輸入目標後點擊「開始測試」，即可同屏對比全部在線節點。',
        diag_current_node: '目前節點：',
        diag_ready: '選擇上方節點並啟動即時串流 MTR 或 Ping 診斷。',
        stat_avg: '平均 RTT',
        stat_min_max: '最小 / 最大',
        stat_jitter: '抖動 (Jitter)',
        stat_loss: '掉包率',
        detected_route: '識別線路類型：',
        col_hop_host: '跳點 IP / PTR',
        col_asn_org: 'ASN / 電信商 & IXP',
        mtr_empty: '等待開始即時 MTR 路由追蹤…',
        speedtest_title: 'HTML5 瀏覽器直連測速',
        speedtest_desc: '純記憶體偽隨機不可壓縮流直連測試，零磁碟落盤開銷。',
        start_speedtest: '開始測速',
        dl_mbps: '下載頻寬 (Mbps)',
        ul_mbps: '上傳頻寬 (Mbps)',
        testfiles_title: '零落盤測速檔案下載',
        testfiles_desc: '伺服器記憶體即時串流產生，支援瀏覽器或指令列拉取。',
        iperf_title: 'iPerf3 臨時連接埠租約',
        iperf_desc: '按需分配獨佔臨時連接埠並啟動單工作階段 iPerf3 伺服器，到期自動銷毀。',
        lease_iperf: '申請臨時 iPerf3 連接埠',
        lease_port: '分配連接埠：',
        copy: '複製',
        copied: '已複製!',
        share_title: '論壇直貼 Markdown 報告匯出',
        share_desc: '自動彙整最近一次橫評、MTR、Ping 及測速結果，預設遮罩目標公網 IP 末段。',
        redact_ip: '自動脫敏目標 IP',
        copy_md: '一鍵複製 Markdown',
        view_hops: '查看跳點',
        hide_hops: '收起',
        probing: '探測中...',
        bgp_title: 'BGP 路由表查詢與 AS-Path 拓撲圖',
        bgp_desc: '唯讀對接 BIRD 2/3、FRRouting、GoBGP，支援前綴路由優選、Large Community 解碼與純 SVG 路徑拓撲。',
        bgp_peers_btn: '查看 BGP 鄰居工作階段',
        bgp_disabled: '目前節點未開啟 BGP 模組（可在設定檔啟用 [bgp]）。',
        bgp_raw: '查看路由器原始輸出 (Raw RIB Output)',
    },
    'ja-JP': {
        your_ip: 'あなたの IP',
        probe_me: '自端末を測定',
        nodes_title: '利用可能なノード',
        refresh: '更新',
        tab_compare: '⚡ マルチノード一括比較',
        tab_diag: '🧭 MTR ルート & Ping',
        tab_speed: '🚀 速度測定 & iPerf3',
        tab_bgp: '🌐 BGP & AS-Path',
        tab_share: '📋 Markdown 共有',
        tab_dig: '🔎 DNS Dig',
        target_label: '対象 IP またはホスト名',
        ip_version: 'IP スタック',
        diag_tool: '診断ツール',
        dig_rrtype: 'レコードタイプ',
        dig_resolver: 'リゾルバー',
        dig_title: 'DNS クエリ (dig)',
        dig_desc: 'ノード視点での完全な dig セマンティクス：レコードごとの TTL・CNAME チェーン・完全な応答セクションを返します。空欄でシステムリゾルバー、指定でパブリック DNS 比較。',
        dig_empty: 'ドメイン（PTR は IP 直接入力可）を入力して「測定開始」を押してください。',
        dig_col_name: '名前',
        dig_col_value: '値',
        tab_http: '🔗 HTTP プローブ',
        http_title: 'HTTP/HTTPS 到達性プローブ',
        http_desc: 'ノードから制御付き GET：ステータスコード・TTFB・リダイレクトチェーン・TLS 証明書サマリー・全レスポンスヘッダー。ボディは 1MiB の計測上限のみで表示はされません。',
        http_empty: '上に完全な URL（例: https://example.com）を入力して「測定開始」を押してください。スキーム省略時は https:// になります。',
        http_status: 'ステータス',
        http_total: '合計 (ms)',
        http_bytes: '読み取り Body',
        http_hop_url: 'ホップ URL',
        http_col_header: 'ヘッダー',
        http_col_value: '値',
        http_tls: 'TLS 証明書',
        http_truncated: '切り詰め（>1MiB）',
        rpki_title: 'RPKI ROA 検証',
        rpki_no_roa: 'このプレフィックス+Origin をカバーする ROA はありません（NotFound）',
        tab_unlock: '📺 ストリーミング判定',
        unlock_title: 'ストリーミングアンロック判定',
        unlock_desc: 'このノードの出口から固定 5 サービスの可用性を probing（公開・ログイン不要エンドポイント、純 Go）。結果はノード単位でキャッシュ。',
        unlock_empty: '「判定実行」で YouTube / Netflix / Disney+ / ChatGPT / TikTok をこのノードから判定します。',
        unlock_run: '判定実行',
        unlock_cache_age: 'キャッシュ',
        unlock_service: 'サービス',
        unlock_status: 'ステータス',
        unlock_region: '地域',
        start_test: '測定開始',
        stop_test: '停止',
        compare_heading: '全リージョン同時レイテンシ・復路ルート比較',
        compare_desc: '全オンライン拠点から同時に RTT・ロス率・復路バックボーン / IXP を自動判定して比較します。',
        include_route: '復路ルートも追跡',
        col_node: 'ノード',
        col_route: '復路キャリア判定',
        col_avg_rtt: '平均遅延',
        col_minmax: '最小 / 最大 / ジッター',
        col_loss: 'ロス率',
        col_detail: 'ホップ',
        compare_empty: '「自端末を測定」または対象を入力して「測定開始」を押すと全ノードを一括比較します。',
        diag_current_node: '選択中ノード：',
        diag_ready: '上のノードを選択してリアルタイム MTR または Ping を開始します。',
        stat_avg: '平均 RTT',
        stat_min_max: '最小 / 最大',
        stat_jitter: 'ジッター',
        stat_loss: 'パケットロス',
        detected_route: '判定ルート：',
        col_hop_host: 'ホップ IP / PTR',
        col_asn_org: 'ASN / 事業者 & IXP',
        mtr_empty: 'リアルタイム MTR トレース待機中…',
        speedtest_title: 'HTML5 ブラウザ速度測定',
        speedtest_desc: 'メモリ上の非圧縮 ChaCha8 ストリームによるディスク負荷ゼロの速度測定。',
        start_speedtest: '速度測定を開始',
        dl_mbps: 'ダウンロード (Mbps)',
        ul_mbps: 'アップロード (Mbps)',
        testfiles_title: 'ゼロディスク・テストファイル',
        testfiles_desc: 'サーバーメモリから直接ストリーミング生成されます。',
        iperf_title: '一時 iPerf3 ポートリース',
        iperf_desc: '専用の一時ポートを割り当て、期限切れで自動停止する iPerf3 サーバーを起動します。',
        lease_iperf: '一時 iPerf3 ポートを取得',
        lease_port: '割当ポート：',
        copy: 'コピー',
        copied: 'コピー完了!',
        share_title: 'マスク済み Markdown レポート出力',
        share_desc: '比較・MTR・Ping・速度測定の結果をターゲット IP 末尾マスク付きで出力します。',
        redact_ip: '対象 IP 末尾をマスク',
        copy_md: 'Markdown をコピー',
        view_hops: 'ホップ表示',
        hide_hops: '閉じる',
        probing: '測定中...',
        bgp_title: 'BGP RIB 検索 & AS-Path トポロジ図',
        bgp_desc: 'BIRD 2/3・FRRouting・GoBGP 読み取り専用連携と純 SVG トポロジ描画。',
        bgp_peers_btn: 'BGP ピア一覧を表示',
        bgp_disabled: 'このノードでは BGP モジュールが無効です（config.toml の [bgp] で有効化可能）。',
        bgp_raw: 'ルーター生出力を表示 (Raw RIB Output)',
    }
};

const state = {
    locale: localStorage.getItem('vmlg-locale') || 'en-US',
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
    const dict = I18N[state.locale] || I18N['en-US'];
    return dict[key] || I18N['en-US'][key] || key;
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
    const sel = document.getElementById('lang-select');
    if (sel && I18N[state.locale]) {
        sel.value = state.locale;
    }
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

    // Streaming unlock panel is opt-in: the tab only exists when the hub
    // advertises unlock_enabled (set [unlock] enabled = true).
    const unlockTab = document.getElementById('tab-btn-unlock');
    if (unlockTab) unlockTab.hidden = !meta.unlock_enabled;

    // Populate footer attribution & community links
    const links = meta.links || {};
    const repoLink = document.getElementById('footer-repo-link');
    if (repoLink && links.repo_url) {
        repoLink.href = links.repo_url;
    }
    const commWrap = document.getElementById('footer-community-wrap');
    const commLink = document.getElementById('footer-community-link');
    if (commWrap && commLink) {
        if (links.community_url) {
            commLink.href = links.community_url;
            commLink.textContent = links.community_label || 'Community';
            commWrap.hidden = false;
        } else {
            commWrap.hidden = true;
        }
    }
    const brandWrap = document.getElementById('footer-brand-wrap');
    const brandLink = document.getElementById('footer-brand-link');
    if (brandWrap && brandLink && links.brand_url) {
        brandLink.href = links.brand_url;
        brandLink.textContent = links.brand_label || meta.title || 'Home';
        brandWrap.hidden = false;
    }

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
    if (!preserveSelection) {
        applyInitialURLParams();
    }
    renderNodes();
    syncSelectedNodeUI();
}

function applyInitialURLParams() {
    const qs = new URLSearchParams(window.location.search);
    if (qs.get('embed') === '1') {
        document.body.classList.add('embed-mode');
        const fullWrap = document.getElementById('footer-open-full-wrap');
        const fullLink = document.getElementById('footer-open-full');
        if (fullWrap && fullLink) {
            const cleanQS = new URLSearchParams(window.location.search);
            cleanQS.delete('embed');
            const query = cleanQS.toString();
            fullLink.href = window.location.pathname + (query ? `?${query}` : '');
            fullWrap.hidden = false;
        }
    }

    const qNode = qs.get('node');
    if (qNode && state.nodes.some((n) => n.id === qNode)) {
        state.selectedNodeId = qNode;
    }
    const qTab = qs.get('tab');
    if (qTab && ['compare', 'diag', 'dig', 'http', 'speed', 'bgp', 'share'].includes(qTab)) {
        switchTab(qTab);
    }
    const qIPVer = qs.get('ip_version');
    if (qIPVer && ['auto', 'v4', 'v6'].includes(qIPVer)) {
        document.getElementById('ipver-select').value = qIPVer;
    }
    const qTool = qs.get('tool');
    if (qTool && ['mtr', 'icmp', 'tcp'].includes(qTool)) {
        document.getElementById('diag-tool-select').value = qTool;
    }
    const qRRType = qs.get('rrtype');
    if (qRRType && ['A', 'AAAA', 'CNAME', 'MX', 'NS', 'TXT', 'PTR'].includes(qRRType)) {
        document.getElementById('dig-rrtype-select').value = qRRType;
    }
    const qResolver = qs.get('resolver');
    if (qResolver) {
        document.getElementById('dig-resolver-input').value = qResolver;
    }
    const qTarget = qs.get('target');
    if (qTarget) {
        if (qTarget === 'me' && state.meta && state.meta.visitor && state.meta.visitor.ip) {
            document.getElementById('target-input').value = state.meta.visitor.ip;
        } else {
            document.getElementById('target-input').value = qTarget;
        }
    }

    if (qs.get('autorun') === '1') {
        setTimeout(() => {
            const target = document.getElementById('target-input').value.trim();
            if (target && (state.activeTab === 'compare' || state.activeTab === 'diag' || state.activeTab === 'dig' || state.activeTab === 'http' || state.activeTab === 'bgp')) {
                document.getElementById('probe-form').requestSubmit();
            }
        }, 80);
    }
}

function buildDeepLinkURL(includeEmbed = false) {
    const u = new URL(window.location.origin + window.location.pathname);
    u.searchParams.set('tab', state.activeTab);
    if (state.selectedNodeId) {
        u.searchParams.set('node', state.selectedNodeId);
    }
    const target = document.getElementById('target-input').value.trim();
    if (target && (state.activeTab === 'compare' || state.activeTab === 'diag' || state.activeTab === 'dig' || state.activeTab === 'http' || state.activeTab === 'bgp')) {
        const visitorIP = state.meta && state.meta.visitor && state.meta.visitor.ip;
        u.searchParams.set('target', target === visitorIP ? 'me' : target);
        u.searchParams.set('autorun', '1');
    }
    const ipVer = document.getElementById('ipver-select').value;
    if (ipVer && ipVer !== 'auto') {
        u.searchParams.set('ip_version', ipVer);
    }
    if (state.activeTab === 'diag') {
        u.searchParams.set('tool', document.getElementById('diag-tool-select').value);
    }
    if (state.activeTab === 'dig') {
        u.searchParams.set('rrtype', document.getElementById('dig-rrtype-select').value);
        const resolver = document.getElementById('dig-resolver-input').value.trim();
        if (resolver) u.searchParams.set('resolver', resolver);
    }
    if (includeEmbed) {
        u.searchParams.set('embed', '1');
        u.searchParams.delete('autorun');
    }
    return u.toString();
}

// Opt-in streaming unlock check: fixed five services from the selected node.
async function runUnlockCheck() {
    const node = selectedNode();
    if (!node) return;
    const btn = document.getElementById('btn-run-unlock');
    btn.disabled = true;
    document.getElementById('unlock-cache-age').textContent = '';
    try {
        const res = await fetch(`/api/v1/unlock?node=${encodeURIComponent(node.id)}`);
        const data = await res.json();
        if (!res.ok) {
            document.getElementById('unlock-tbody').innerHTML =
                `<tr class="empty-row"><td colspan="4">${escapeHTML(data.error || 'unlock check failed')}</td></tr>`;
            return;
        }
        renderUnlockResult(data);
    } catch (e) {
        document.getElementById('unlock-tbody').innerHTML =
            `<tr class="empty-row"><td colspan="4">${escapeHTML(e.message)}</td></tr>`;
    } finally {
        btn.disabled = false;
    }
}

function renderUnlockResult(data) {
    const age = document.getElementById('unlock-cache-age');
    age.textContent = data.cached ? `${t('unlock_cache_age')}: ${data.cache_age_seconds}s` : '';

    const rows = (data.results || []).map((r) => {
        let badgeClass = 'line-badge subtle';
        let label = (r.status || '').toUpperCase();
        if (r.status === 'unlocked' || r.status === 'available') badgeClass = 'line-badge premium';
        if (r.status === 'locked' || r.status === 'unavailable') {
            badgeClass = 'line-badge';
            label += r.error ? '' : '';
        }
        const title = r.error ? ` title="${escapeHTML(r.error)}"` : '';
        return `
            <tr>
                <td>${escapeHTML(r.service)}</td>
                <td><span class="${badgeClass}"${title}>${escapeHTML(label)}</span></td>
                <td class="mono">${escapeHTML(r.region || '—')}</td>
                <td class="mono">${escapeHTML(String(r.elapsed_ms ?? '—'))}</td>
            </tr>
        `;
    }).join('');
    document.getElementById('unlock-tbody').innerHTML = rows ||
        `<tr class="empty-row"><td colspan="4">—</td></tr>`;
}

function fmtUptime(seconds) {
    const s = Number(seconds) || 0;
    if (s < 60) return `${s}s`;
    const m = Math.floor(s / 60);
    if (m < 60) return `${m}m`;
    const h = Math.floor(m / 60);
    if (h < 24) return `${h}h${m % 60}m`;
    const d = Math.floor(h / 24);
    return `${d}d${h % 24}h`;
}

// 1h in-memory load sparkline (240 × 15s points from /api/v1/nodes).
function sparklineSVG(samples) {
    const w = 220;
    const h = 26;
    const loads = samples.map((p) => p.load1 || 0);
    const max = Math.max(...loads, 0.01);
    const step = w / Math.max(loads.length - 1, 1);
    const points = loads.map((v, i) => `${(i * step).toFixed(1)},${(h - (v / max) * (h - 2) - 1).toFixed(1)}`).join(' ');
    return `<svg viewBox="0 0 ${w} ${h}" preserveAspectRatio="none" role="img" aria-label="1h load">
        <polyline points="${points}" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.85"/>
    </svg>`;
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
        const health = n.online && n.loadavg_1 !== undefined && n.loadavg_1 !== null
            ? `<span title="load 1/5/15 · mem · uptime">⚡ ${Number(n.loadavg_1).toFixed(2)} / ${Number(n.loadavg_5).toFixed(2)} / ${Number(n.loadavg_15).toFixed(2)} · 💾 ${Math.round(n.mem_used_percent || 0)}% · ⏱ ${fmtUptime(n.uptime_seconds)}</span>`
            : '';
        const spark = (n.online && n.load_samples && n.load_samples.length > 3) ? sparklineSVG(n.load_samples) : '';
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
                ${health ? `<div class="node-health mono">${health}</div>` : ''}
                ${spark ? `<div class="node-spark">${spark}</div>` : ''}
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
    document.getElementById('dig-node-label').textContent = label;
    document.getElementById('http-node-label').textContent = label;
    document.getElementById('unlock-node-label').textContent = label;

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
    ['compare', 'diag', 'dig', 'http', 'speed', 'bgp', 'unlock', 'share'].forEach((id) => {
        const panel = document.getElementById(`panel-${id}`);
        if (panel) panel.hidden = id !== tabName;
    });
    const showTargetBar = tabName === 'compare' || tabName === 'diag' || tabName === 'dig' || tabName === 'http' || tabName === 'bgp';
    document.getElementById('target-bar').hidden = !showTargetBar;
    document.getElementById('diag-tool-group').hidden = tabName !== 'diag';
    document.getElementById('dig-rrtype-group').hidden = tabName !== 'dig';
    document.getElementById('dig-resolver-group').hidden = tabName !== 'dig';
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
async function startDig(target) {
    const node = selectedNode();
    if (!node) return;
    showProbeError('');
    setProbeRunning(true);
    state.lastTarget = target;

    try {
        const qs = new URLSearchParams({
            node: node.id,
            target,
            rrtype: document.getElementById('dig-rrtype-select').value,
        });
        const resolver = document.getElementById('dig-resolver-input').value.trim();
        if (resolver) qs.set('resolver', resolver);

        const res = await fetch(`/api/v1/probe/dig?${qs.toString()}`);
        const data = await res.json();
        if (!res.ok) {
            showProbeError(data.error || 'dig query failed');
            return;
        }
        renderDigResult(data);
    } catch (e) {
        showProbeError(e.message);
    } finally {
        setProbeRunning(false);
    }
}

function renderDigResult(res) {
    state.lastDig = res;

    const pill = document.getElementById('dig-meta-pill');
    pill.hidden = false;
    const elapsed = typeof res.elapsed_ms === 'number' ? res.elapsed_ms.toFixed(2) : res.elapsed_ms;
    pill.textContent = `${res.rcode} · ${res.resolver} · ${String(res.transport || 'udp').toUpperCase()} · ${elapsed} ms`;

    const tbody = document.getElementById('dig-tbody');
    if (!res.answers || res.answers.length === 0) {
        tbody.innerHTML = `<tr class="empty-row"><td colspan="4">${escapeHTML(res.rcode)}</td></tr>`;
        return;
    }
    tbody.innerHTML = res.answers.map((a) => `
        <tr>
            <td class="mono">${escapeHTML(a.name)}</td>
            <td class="mono">${escapeHTML(String(a.ttl))}</td>
            <td><span class="tag-chip">${escapeHTML(a.type)}</span></td>
            <td class="mono">${escapeHTML(a.value)}</td>
        </tr>
    `).join('');
}

async function startHTTPProbe(target) {
    const node = selectedNode();
    if (!node) return;
    if (!/^https?:\/\//i.test(target)) {
        target = `https://${target}`;
    }
    showProbeError('');
    setProbeRunning(true);
    state.lastTarget = target;

    try {
        const qs = new URLSearchParams({ node: node.id, url: target });
        const res = await fetch(`/api/v1/probe/http?${qs.toString()}`);
        const data = await res.json();
        if (!res.ok) {
            showProbeError(data.error || 'http probe failed');
            return;
        }
        renderHTTPResult(data);
    } catch (e) {
        showProbeError(e.message);
    } finally {
        setProbeRunning(false);
    }
}

function renderHTTPResult(res) {
    state.lastHTTP = res;

    document.getElementById('http-stats-wrap').hidden = false;
    document.getElementById('http-empty-hint').hidden = true;
    document.getElementById('http-status').textContent = `${res.status_code} ${res.status || ''}`.trim();
    document.getElementById('http-ttfb').textContent = fmtMs(res.ttfb_ms);
    document.getElementById('http-total').textContent = fmtMs(res.total_ms);
    const bytes = `${(res.bytes_read / 1024).toFixed(1)} KiB${res.truncated ? ` · ${t('http_truncated')}` : ''}`;
    document.getElementById('http-bytes').textContent = bytes;

    const pill = document.getElementById('http-meta-pill');
    pill.hidden = false;
    pill.textContent = `${res.proto || 'HTTP'} · ${res.remote_addr || ''}`;

    const tlsBox = document.getElementById('http-tls-box');
    if (res.tls) {
        tlsBox.hidden = false;
        document.getElementById('http-tls-issuer').textContent = res.tls.issuer || '—';
        document.getElementById('http-tls-validity').textContent = `${res.tls.not_before} → ${res.tls.not_after}`;
        document.getElementById('http-tls-sans').textContent = (res.tls.sans || []).slice(0, 6).join(', ');
    } else {
        tlsBox.hidden = true;
    }

    const redirWrap = document.getElementById('http-redirects-wrap');
    const redirBody = document.getElementById('http-redirects-tbody');
    if (res.redirects && res.redirects.length > 0) {
        redirWrap.hidden = false;
        redirBody.innerHTML = res.redirects.map((h, i) => `
            <tr>
                <td>${i + 1}</td>
                <td class="mono">${escapeHTML(h.url)}</td>
                <td><span class="tag-chip">${escapeHTML(String(h.status))}</span></td>
                <td class="mono">${escapeHTML(h.location)}</td>
                <td class="mono">${escapeHTML(String(h.elapsed_ms))}</td>
            </tr>
        `).join('');
    } else {
        redirWrap.hidden = true;
    }

    const hdrWrap = document.getElementById('http-headers-wrap');
    const hdrBody = document.getElementById('http-headers-tbody');
    const headers = res.headers || {};
    const hdrRows = Object.keys(headers).sort().map((k) => `
        <tr>
            <td class="mono">${escapeHTML(k)}</td>
            <td class="mono">${escapeHTML((headers[k] || []).join('; '))}</td>
        </tr>
    `).join('');
    if (hdrRows) {
        hdrWrap.hidden = false;
        hdrBody.innerHTML = hdrRows;
    } else {
        hdrWrap.hidden = true;
    }
}

function fmtMs(v) {
    return typeof v === 'number' ? v.toFixed(2) : String(v);
}

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

// RPKI validity badge for the best path (only present when routinator is configured).
function renderRPKIBadge(rpki) {
    const banner = document.getElementById('bgp-rpki-banner');
    const badge = document.getElementById('bgp-rpki-badge');
    const detail = document.getElementById('bgp-rpki-detail');
    if (!rpki || !rpki.state) {
        banner.hidden = true;
        return;
    }
    banner.hidden = false;
    const vrpText = (rpki.vrps || [])
        .map((v) => `${v.asn} · ${v.prefix} · max /${v.max_length || '?'}`)
        .join('  |  ');
    switch (rpki.state) {
        case 'valid':
            badge.textContent = 'VALID';
            badge.className = 'line-badge premium';
            detail.textContent = vrpText;
            break;
        case 'invalid':
            badge.textContent = 'INVALID';
            badge.className = 'line-badge';
            badge.style.cssText = 'background:#dc2626;color:#fff';
            detail.textContent = `${rpki.reason ? escapeHTML(rpki.reason) + ' — ' : ''}${escapeHTML(vrpText)}`;
            break;
        case 'notfound':
            badge.textContent = 'NOT FOUND';
            badge.className = 'line-badge subtle';
            detail.textContent = t('rpki_no_roa');
            break;
        default: // unavailable
            badge.textContent = 'UNAVAILABLE';
            badge.className = 'line-badge subtle';
            detail.textContent = escapeHTML(rpki.reason || '');
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

    renderRPKIBadge(data.rpki);

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

// Share Markdown & BBCode Builder
function renderShareMarkdown() {
    const fmtEl = document.getElementById('share-format-select');
    const format = fmtEl ? fmtEl.value : 'markdown';
    const redact = document.getElementById('share-redact-toggle').checked;
    const fmtTarget = (raw) => (redact ? redactIP(raw) : raw);
    const repoURL = (state.meta && state.meta.links && state.meta.links.repo_url) || 'https://github.com/cloudapp3/vmlg';
    const title = (state.meta && state.meta.title) || 'vmlg Looking Glass';
    const cmpItems = Object.values(state.compareResults).filter((x) => x.ping || x.mtr);

    if (format === 'bbcode') {
        document.getElementById('share-markdown-box').value = buildBBCodeReport(title, repoURL, fmtTarget, redact, cmpItems);
        return;
    }

    const lines = [];
    lines.push(`### ${title} - Network Diagnostic Report`);
    lines.push(`- Generated: \`${new Date().toISOString()}\``);
    if (state.lastTarget) {
        lines.push(`- Target: \`${fmtTarget(state.lastTarget)}\``);
    }
    lines.push('');

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
        lines.push('_Run a Multi-Node Compare, MTR, Ping, BGP Lookup, or Speedtest first to populate this report._');
        lines.push('');
    }

    lines.push(`> Generated by [vmlg Looking Glass](${repoURL}) — Single-Binary Multi-Node Network & BGP Diagnostics`);
    document.getElementById('share-markdown-box').value = lines.join('\n');
}

function buildBBCodeReport(title, repoURL, fmtTarget, redact, cmpItems) {
    const lines = [];
    lines.push(`[b][size=4]${title} - Network Diagnostic Report[/size][/b]`);
    lines.push(`Generated: [font=monospace]${new Date().toISOString()}[/font]`);
    if (state.lastTarget) {
        lines.push(`Target: [font=monospace]${fmtTarget(state.lastTarget)}[/font]`);
    }
    lines.push('');

    if (cmpItems.length > 0) {
        lines.push('[b]Multi-Node Latency & Return Route[/b]');
        lines.push('[table]');
        lines.push('[tr][td][b]Node[/b][/td][td][b]Region[/b][/td][td][b]Return Route[/b][/td][td][b]Avg RTT[/b][/td][td][b]Loss[/b][/td][/tr]');
        cmpItems.forEach((it) => {
            const route = (it.mtr && it.mtr.route && it.mtr.route.label) || '—';
            const avg = it.ping && it.ping.recv > 0 ? `${it.ping.avg_ms.toFixed(1)} ms` : 'Timeout';
            const loss = it.ping ? `${it.ping.loss_pct.toFixed(0)}%` : '—';
            lines.push(`[tr][td]${it.node.name}[/td][td]${it.node.region || '—'}[/td][td][b]${route}[/b][/td][td]${avg}[/td][td]${loss}[/td][/tr]`);
        });
        lines.push('[/table]');
        lines.push('');
    }

    if (state.lastMTR && state.lastMTR.snap) {
        const { node, snap } = state.lastMTR;
        lines.push(`[b]MTR Trace from ${node.name} (${snap.route ? snap.route.label : 'Direct'})[/b]`);
        lines.push('[code]');
        (snap.hops || []).forEach((h, idx, arr) => {
            const isLast = idx === arr.length - 1;
            const ipDisp = (h.addr ? (isLast && redact ? redactIP(h.addr) : h.addr) : '*').padEnd(18, ' ');
            const tag = ([h.ixp || h.line_tag, h.asn].filter(Boolean).join(' ') || '-').padEnd(18, ' ');
            const avg = h.recv > 0 ? `${h.avg_ms.toFixed(1)}ms` : 'timeout';
            lines.push(`${String(h.hop).padStart(2, ' ')}. ${ipDisp} ${tag} loss=${h.loss_pct.toFixed(0)}% avg=${avg}`);
        });
        lines.push('[/code]');
        lines.push('');
    }

    if (state.lastSpeed) {
        const s = state.lastSpeed;
        lines.push(`[b]HTML5 Speedtest (${s.node.name})[/b]: Download [b]${s.dlMbps} Mbps[/b] · Upload [b]${s.ulMbps} Mbps[/b] · Ping ${s.pingMs} ms (Jitter ${s.jitterMs} ms)`);
        lines.push('');
    }

    lines.push(`[quote]Generated by [url=${repoURL}][b]vmlg Looking Glass[/b][/url] — Single-Binary Multi-Node Network & BGP Portal[/quote]`);
    return lines.join('\n');
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

    document.getElementById('lang-select').addEventListener('change', (e) => {
        state.locale = e.target.value;
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
        } else if (state.activeTab === 'dig') {
            startDig(target);
        } else if (state.activeTab === 'http') {
            startHTTPProbe(target);
        } else {
            startSingleDiag(target, ipVer);
        }
    });

    document.getElementById('btn-bgp-summary').addEventListener('click', loadBGPSummary);

    document.getElementById('btn-run-unlock').addEventListener('click', runUnlockCheck);

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
    document.getElementById('share-format-select').addEventListener('change', renderShareMarkdown);
    document.getElementById('btn-copy-markdown').addEventListener('click', (e) => {
        copyText(document.getElementById('share-markdown-box').value, e.currentTarget);
    });
    document.getElementById('btn-copy-deeplink').addEventListener('click', (e) => {
        copyText(buildDeepLinkURL(false), e.currentTarget);
    });
    document.getElementById('btn-copy-embed').addEventListener('click', (e) => {
        const embedURL = buildDeepLinkURL(true);
        const snippet = `<iframe src="${embedURL}" width="100%" height="540" style="border:0;border-radius:12px;overflow:hidden" loading="lazy"></iframe>`;
        copyText(snippet, e.currentTarget);
    });
}

bindEvents();
applyLocale();
loadMeta(false);
