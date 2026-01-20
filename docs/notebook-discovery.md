# Notebook Discovery

OpenNotes automatically discovers and loads notebooks based on the user's current working directory using a sophisticated 3-tier priority system. This document outlines the complete algorithm and provides a visual flowchart of the discovery process.

## Overview

The notebook discovery follows a **3-tier priority system**:

1. **Declared Path** (highest priority) - From `--notebook` flag or `OPENNOTES_NOTEBOOK` env var
2. **Registered Notebooks** (medium priority) - Check each registered notebook for context match
3. **Ancestor Search** (fallback) - Walk up directory tree looking for `.opennotes.json`

## Discovery Flowchart

<?xml version="1.0" encoding="utf-8"?><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" data-d2-version="v0.7.1" preserveAspectRatio="xMinYMin meet" viewBox="0 0 1320 3300"><svg class="d2-3094856511 d2-svg" width="1320" height="3300" viewBox="-101 -101 1320 3300"><rect x="-101.000000" y="-101.000000" width="1320.000000" height="3300.000000" rx="0.000000" fill="#FFFFFF" class=" fill-N7" stroke-width="0" /><style type="text/css"><![CDATA[
.d2-3094856511 .text-bold {
	font-family: "d2-3094856511-font-bold";
}
@font-face {
	font-family: d2-3094856511-font-bold;
	src: url("data:application/font-woff;base64,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");
}
.d2-3094856511 .text-italic {
	font-family: "d2-3094856511-font-italic";
}
@font-face {
	font-family: d2-3094856511-font-italic;
	src: url("data:application/font-woff;base64,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");
}]]></style><style type="text/css"><![CDATA[.shape {
  shape-rendering: geometricPrecision;
  stroke-linejoin: round;
}
.connection {
  stroke-linecap: round;
  stroke-linejoin: round;
}
.blend {
  mix-blend-mode: multiply;
  opacity: 0.5;
}

		.d2-3094856511 .fill-N1{fill:#0A0F25;}
		.d2-3094856511 .fill-N2{fill:#676C7E;}
		.d2-3094856511 .fill-N3{fill:#9499AB;}
		.d2-3094856511 .fill-N4{fill:#CFD2DD;}
		.d2-3094856511 .fill-N5{fill:#DEE1EB;}
		.d2-3094856511 .fill-N6{fill:#EEF1F8;}
		.d2-3094856511 .fill-N7{fill:#FFFFFF;}
		.d2-3094856511 .fill-B1{fill:#0D32B2;}
		.d2-3094856511 .fill-B2{fill:#0D32B2;}
		.d2-3094856511 .fill-B3{fill:#E3E9FD;}
		.d2-3094856511 .fill-B4{fill:#E3E9FD;}
		.d2-3094856511 .fill-B5{fill:#EDF0FD;}
		.d2-3094856511 .fill-B6{fill:#F7F8FE;}
		.d2-3094856511 .fill-AA2{fill:#4A6FF3;}
		.d2-3094856511 .fill-AA4{fill:#EDF0FD;}
		.d2-3094856511 .fill-AA5{fill:#F7F8FE;}
		.d2-3094856511 .fill-AB4{fill:#EDF0FD;}
		.d2-3094856511 .fill-AB5{fill:#F7F8FE;}
		.d2-3094856511 .stroke-N1{stroke:#0A0F25;}
		.d2-3094856511 .stroke-N2{stroke:#676C7E;}
		.d2-3094856511 .stroke-N3{stroke:#9499AB;}
		.d2-3094856511 .stroke-N4{stroke:#CFD2DD;}
		.d2-3094856511 .stroke-N5{stroke:#DEE1EB;}
		.d2-3094856511 .stroke-N6{stroke:#EEF1F8;}
		.d2-3094856511 .stroke-N7{stroke:#FFFFFF;}
		.d2-3094856511 .stroke-B1{stroke:#0D32B2;}
		.d2-3094856511 .stroke-B2{stroke:#0D32B2;}
		.d2-3094856511 .stroke-B3{stroke:#E3E9FD;}
		.d2-3094856511 .stroke-B4{stroke:#E3E9FD;}
		.d2-3094856511 .stroke-B5{stroke:#EDF0FD;}
		.d2-3094856511 .stroke-B6{stroke:#F7F8FE;}
		.d2-3094856511 .stroke-AA2{stroke:#4A6FF3;}
		.d2-3094856511 .stroke-AA4{stroke:#EDF0FD;}
		.d2-3094856511 .stroke-AA5{stroke:#F7F8FE;}
		.d2-3094856511 .stroke-AB4{stroke:#EDF0FD;}
		.d2-3094856511 .stroke-AB5{stroke:#F7F8FE;}
		.d2-3094856511 .background-color-N1{background-color:#0A0F25;}
		.d2-3094856511 .background-color-N2{background-color:#676C7E;}
		.d2-3094856511 .background-color-N3{background-color:#9499AB;}
		.d2-3094856511 .background-color-N4{background-color:#CFD2DD;}
		.d2-3094856511 .background-color-N5{background-color:#DEE1EB;}
		.d2-3094856511 .background-color-N6{background-color:#EEF1F8;}
		.d2-3094856511 .background-color-N7{background-color:#FFFFFF;}
		.d2-3094856511 .background-color-B1{background-color:#0D32B2;}
		.d2-3094856511 .background-color-B2{background-color:#0D32B2;}
		.d2-3094856511 .background-color-B3{background-color:#E3E9FD;}
		.d2-3094856511 .background-color-B4{background-color:#E3E9FD;}
		.d2-3094856511 .background-color-B5{background-color:#EDF0FD;}
		.d2-3094856511 .background-color-B6{background-color:#F7F8FE;}
		.d2-3094856511 .background-color-AA2{background-color:#4A6FF3;}
		.d2-3094856511 .background-color-AA4{background-color:#EDF0FD;}
		.d2-3094856511 .background-color-AA5{background-color:#F7F8FE;}
		.d2-3094856511 .background-color-AB4{background-color:#EDF0FD;}
		.d2-3094856511 .background-color-AB5{background-color:#F7F8FE;}
		.d2-3094856511 .color-N1{color:#0A0F25;}
		.d2-3094856511 .color-N2{color:#676C7E;}
		.d2-3094856511 .color-N3{color:#9499AB;}
		.d2-3094856511 .color-N4{color:#CFD2DD;}
		.d2-3094856511 .color-N5{color:#DEE1EB;}
		.d2-3094856511 .color-N6{color:#EEF1F8;}
		.d2-3094856511 .color-N7{color:#FFFFFF;}
		.d2-3094856511 .color-B1{color:#0D32B2;}
		.d2-3094856511 .color-B2{color:#0D32B2;}
		.d2-3094856511 .color-B3{color:#E3E9FD;}
		.d2-3094856511 .color-B4{color:#E3E9FD;}
		.d2-3094856511 .color-B5{color:#EDF0FD;}
		.d2-3094856511 .color-B6{color:#F7F8FE;}
		.d2-3094856511 .color-AA2{color:#4A6FF3;}
		.d2-3094856511 .color-AA4{color:#EDF0FD;}
		.d2-3094856511 .color-AA5{color:#F7F8FE;}
		.d2-3094856511 .color-AB4{color:#EDF0FD;}
		.d2-3094856511 .color-AB5{color:#F7F8FE;}.appendix text.text{fill:#0A0F25}.md{--color-fg-default:#0A0F25;--color-fg-muted:#676C7E;--color-fg-subtle:#9499AB;--color-canvas-default:#FFFFFF;--color-canvas-subtle:#EEF1F8;--color-border-default:#0D32B2;--color-border-muted:#0D32B2;--color-neutral-muted:#EEF1F8;--color-accent-fg:#0D32B2;--color-accent-emphasis:#0D32B2;--color-attention-subtle:#676C7E;--color-danger-fg:red;}.sketch-overlay-B1{fill:url(#streaks-darker-d2-3094856511);mix-blend-mode:lighten}.sketch-overlay-B2{fill:url(#streaks-darker-d2-3094856511);mix-blend-mode:lighten}.sketch-overlay-B3{fill:url(#streaks-bright-d2-3094856511);mix-blend-mode:darken}.sketch-overlay-B4{fill:url(#streaks-bright-d2-3094856511);mix-blend-mode:darken}.sketch-overlay-B5{fill:url(#streaks-bright-d2-3094856511);mix-blend-mode:darken}.sketch-overlay-B6{fill:url(#streaks-bright-d2-3094856511);mix-blend-mode:darken}.sketch-overlay-AA2{fill:url(#streaks-dark-d2-3094856511);mix-blend-mode:overlay}.sketch-overlay-AA4{fill:url(#streaks-bright-d2-3094856511);mix-blend-mode:darken}.sketch-overlay-AA5{fill:url(#streaks-bright-d2-3094856511);mix-blend-mode:darken}.sketch-overlay-AB4{fill:url(#streaks-bright-d2-3094856511);mix-blend-mode:darken}.sketch-overlay-AB5{fill:url(#streaks-bright-d2-3094856511);mix-blend-mode:darken}.sketch-overlay-N1{fill:url(#streaks-darker-d2-3094856511);mix-blend-mode:lighten}.sketch-overlay-N2{fill:url(#streaks-dark-d2-3094856511);mix-blend-mode:overlay}.sketch-overlay-N3{fill:url(#streaks-normal-d2-3094856511);mix-blend-mode:color-burn}.sketch-overlay-N4{fill:url(#streaks-normal-d2-3094856511);mix-blend-mode:color-burn}.sketch-overlay-N5{fill:url(#streaks-bright-d2-3094856511);mix-blend-mode:darken}.sketch-overlay-N6{fill:url(#streaks-bright-d2-3094856511);mix-blend-mode:darken}.sketch-overlay-N7{fill:url(#streaks-bright-d2-3094856511);mix-blend-mode:darken}.light-code{display: block}.dark-code{display: none}]]></style><g class="U3RhcnQ="><g class="shape" ><ellipse rx="164.500000" ry="55.000000" cx="293.500000" cy="55.000000" stroke="#01579b" fill="#e1f5fe" class="shape" style="stroke-width:2;" /></g><text x="293.500000" y="52.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="293.500000" dy="0.000000">Start:</tspan><tspan x="293.500000" dy="18.500000">Current Working Directory</tspan></text></g><g class="Q2hlY2tEZWNsYXJlZA=="><g class="shape" ><path d="M 293 398 C 291 398 290 398 289 397 L 73 306 C 71 305 71 304 73 303 L 289 211 C 291 210 295 210 297 211 L 513 303 C 515 304 515 305 513 306 L 297 397 C 296 398 295 398 293 398 Z" stroke="#ef6c00" fill="#fff3e0" style="stroke-width:2;" /></g><text x="293.000000" y="285.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="293.000000" dy="0.000000">Check Declared</tspan><tspan x="293.000000" dy="17.250000">Notebook Path</tspan><tspan x="293.000000" dy="17.250000">--notebook flag or</tspan><tspan x="293.000000" dy="17.250000">OPENNOTES_NOTEBOOK env</tspan></text></g><g class="SGFzRGVjbGFyZWRDb25maWc="><g class="shape" ><path d="M 159 643 C 158 643 157 643 156 643 L 1 582 C -1 581 -1 580 1 580 L 156 519 C 158 518 160 518 162 519 L 317 579 C 319 580 319 581 317 581 L 162 643 C 161 643 160 643 159 643 Z" stroke="#ef6c00" fill="#fff3e0" style="stroke-width:2;" /></g><text x="159.000000" y="578.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="159.000000" dy="0.000000">Has .opennotes.json</tspan><tspan x="159.000000" dy="18.500000">in declared path?</tspan></text></g><g class="TG9hZERlY2xhcmVk"><g class="shape" ><rect x="69.000000" y="2807.000000" width="180.000000" height="82.000000" stroke="#4a148c" fill="#f3e5f5" style="stroke-width:2;" /></g><text x="159.000000" y="2845.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="159.000000" dy="0.000000">Load &amp; Open</tspan><tspan x="159.000000" dy="18.500000">Declared Notebook</tspan></text></g><g class="Q2hlY2tSZWdpc3RlcmVk"><g class="shape" ><path d="M 414 920 C 413 920 412 920 411 919 L 277 844 C 276 843 276 842 277 841 L 411 765 C 412 764 415 764 416 765 L 550 841 C 551 842 551 843 550 844 L 417 919 C 416 920 415 920 414 920 Z" stroke="#ef6c00" fill="#fff3e0" style="stroke-width:2;" /></g><text x="414.000000" y="831.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="414.000000" dy="0.000000">Check Registered</tspan><tspan x="414.000000" dy="17.666667">Notebooks from</tspan><tspan x="414.000000" dy="17.666667">global config</tspan></text></g><g class="Rm9yRWFjaFJlZ2lzdGVyZWQ="><g class="shape" ><rect x="323.000000" y="1020.000000" width="181.000000" height="82.000000" stroke="#4a148c" fill="#f3e5f5" style="stroke-width:2;" /></g><text x="413.500000" y="1058.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="413.500000" dy="0.000000">For each registered</tspan><tspan x="413.500000" dy="18.500000">notebook path</tspan></text></g><g class="SGFzUmVnaXN0ZXJlZENvbmZpZw=="><g class="shape" ><path d="M 414 1326 C 413 1326 412 1326 411 1326 L 256 1265 C 254 1264 254 1263 256 1263 L 411 1202 C 413 1201 415 1201 417 1202 L 572 1262 C 574 1263 574 1264 572 1264 L 417 1326 C 416 1326 415 1326 414 1326 Z" stroke="#ef6c00" fill="#fff3e0" style="stroke-width:2;" /></g><text x="414.000000" y="1261.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="414.000000" dy="0.000000">Has .opennotes.json</tspan><tspan x="414.000000" dy="18.500000">in registered path?</tspan></text></g><g class="TG9hZFJlZ2lzdGVyZWRDb25maWc="><g class="shape" ><rect x="395.000000" y="1447.000000" width="198.000000" height="66.000000" stroke="#4a148c" fill="#f3e5f5" style="stroke-width:2;" /></g><text x="494.000000" y="1485.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px">Load notebook config</text></g><g class="Q2hlY2tDb250ZXh0"><g class="shape" ><path d="M 494 1790 C 493 1790 492 1790 491 1789 L 333 1714 C 331 1713 331 1712 333 1711 L 491 1635 C 493 1634 495 1634 497 1635 L 655 1711 C 657 1712 657 1713 655 1714 L 497 1789 C 496 1790 495 1790 494 1790 Z" stroke="#ef6c00" fill="#fff3e0" style="stroke-width:2;" /></g><text x="494.000000" y="1701.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="494.000000" dy="0.000000">Current directory</tspan><tspan x="494.000000" dy="17.666667">matches any context</tspan><tspan x="494.000000" dy="17.666667">in notebook?</tspan></text></g><g class="TG9hZFJlZ2lzdGVyZWQ="><g class="shape" ><rect x="310.000000" y="2807.000000" width="178.000000" height="82.000000" stroke="#4a148c" fill="#f3e5f5" style="stroke-width:2;" /></g><text x="399.000000" y="2845.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="399.000000" dy="0.000000">Load &amp; Open</tspan><tspan x="399.000000" dy="18.500000">Matched Notebook</tspan></text></g><g class="TmV4dFJlZ2lzdGVyZWQ="><g class="shape" ><rect x="725.000000" y="1911.000000" width="190.000000" height="82.000000" stroke="#4a148c" fill="#f3e5f5" style="stroke-width:2;" /></g><text x="820.000000" y="1949.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="820.000000" dy="0.000000">Try next</tspan><tspan x="820.000000" dy="18.500000">registered notebook</tspan></text></g><g class="U3RhcnRBbmNlc3RvclNlYXJjaA=="><g class="shape" ><rect x="720.000000" y="2114.000000" width="199.000000" height="82.000000" stroke="#4a148c" fill="#f3e5f5" style="stroke-width:2;" /></g><text x="819.500000" y="2152.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="819.500000" dy="0.000000">Start Ancestor Search</tspan><tspan x="819.500000" dy="18.500000">current = cwd</tspan></text></g><g class="SXNSb290"><g class="shape" ><path d="M 820 2441 C 819 2441 818 2441 818 2441 L 695 2380 C 694 2379 694 2378 695 2378 L 818 2317 C 819 2316 821 2316 823 2317 L 946 2377 C 947 2378 947 2379 946 2379 L 822 2441 C 822 2441 821 2441 820 2441 Z" stroke="#ef6c00" fill="#fff3e0" style="stroke-width:2;" /></g><text x="820.000000" y="2376.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="820.000000" dy="0.000000">current == &#39;/&#39; or</tspan><tspan x="820.000000" dy="18.500000">empty string?</tspan></text></g><g class="SGFzQW5jZXN0b3JDb25maWc="><g class="shape" ><path d="M 821 2686 C 820 2686 819 2686 818 2686 L 661 2625 C 659 2624 659 2623 661 2623 L 818 2562 C 820 2561 822 2561 824 2562 L 981 2622 C 983 2623 983 2624 981 2624 L 824 2686 C 823 2686 822 2686 821 2686 Z" stroke="#ef6c00" fill="#fff3e0" style="stroke-width:2;" /></g><text x="821.000000" y="2621.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="821.000000" dy="0.000000">Has .opennotes.json</tspan><tspan x="821.000000" dy="18.500000">in current directory?</tspan></text></g><g class="TG9hZEFuY2VzdG9y"><g class="shape" ><rect x="639.000000" y="2807.000000" width="182.000000" height="82.000000" stroke="#4a148c" fill="#f3e5f5" style="stroke-width:2;" /></g><text x="730.000000" y="2845.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="730.000000" dy="0.000000">Load &amp; Open</tspan><tspan x="730.000000" dy="18.500000">Ancestor Notebook</tspan></text></g><g class="R29Ub1BhcmVudA=="><g class="shape" ><rect x="957.000000" y="2807.000000" width="161.000000" height="82.000000" stroke="#4a148c" fill="#f3e5f5" style="stroke-width:2;" /></g><text x="1037.500000" y="2845.500000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="1037.500000" dy="0.000000">current = parent</tspan><tspan x="1037.500000" dy="18.500000">directory</tspan></text></g><g class="U3VjY2Vzcw=="><g class="shape" ><ellipse rx="163.000000" ry="54.500000" cx="378.000000" cy="3043.500000" stroke="#1b5e20" fill="#e8f5e8" class="shape" style="stroke-width:2;" /></g><text x="378.000000" y="3041.000000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="378.000000" dy="0.000000">Success</tspan><tspan x="378.000000" dy="18.500000">Return Notebook Instance</tspan></text></g><g class="Tm90Rm91bmQ="><g class="shape" ><ellipse rx="80.500000" ry="43.500000" cx="519.500000" cy="2624.500000" stroke="#c62828" fill="#ffebee" class="shape" style="stroke-width:2;" /></g><text x="519.500000" y="2622.000000" fill="#0A0F25" class="text-bold fill-N1" style="text-anchor:middle;font-size:16px"><tspan x="519.500000" dy="0.000000">Not Found</tspan><tspan x="519.500000" dy="18.500000">Return nil</tspan></text></g><g class="KFN0YXJ0IC0mZ3Q7IENoZWNrRGVjbGFyZWQpWzBd"><marker id="mk-d2-3094856511-3488378134" markerWidth="10.000000" markerHeight="12.000000" refX="7.000000" refY="6.000000" viewBox="0.000000 0.000000 10.000000 12.000000" orient="auto" markerUnits="userSpaceOnUse"> <polygon points="0.000000,0.000000 10.000000,6.000000 0.000000,12.000000" fill="#0D32B2" class="connection fill-B1" stroke-width="2" /> </marker><path d="M 293.000000 112.000000 C 293.000000 150.000000 293.000000 170.000000 293.000000 206.000000" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /></g><g class="KENoZWNrRGVjbGFyZWQgLSZndDsgSGFzRGVjbGFyZWRDb25maWcpWzBd"><path d="M 230.701350 374.521022 C 173.600006 441.399994 159.000000 470.399994 159.000000 514.000000" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="176.500000" y="443.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">Path provided</text></g><g class="KENoZWNrRGVjbGFyZWQgLSZndDsgQ2hlY2tSZWdpc3RlcmVkKVswXQ=="><path d="M 355.308851 374.512253 C 413.200012 441.399994 428.000000 483.000000 428.000000 519.750000 C 428.000000 556.500000 426.799988 716.400024 422.370493 764.017195" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="428.000000" y="564.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">No path</text></g><g class="KEhhc0RlY2xhcmVkQ29uZmlnIC0mZ3Q7IExvYWREZWNsYXJlZClbMF0="><path d="M 159.000000 645.000000 C 159.000000 691.400024 159.000000 731.200012 159.000000 772.750000 C 159.000000 814.299988 159.000000 867.599976 159.000000 906.000000 C 159.000000 944.400024 159.000000 988.200012 159.000000 1015.500000 C 159.000000 1042.800049 159.000000 1079.199951 159.000000 1106.500000 C 159.000000 1133.800049 159.000000 1174.400024 159.000000 1208.000000 C 159.000000 1241.599976 159.000000 1288.500000 159.000000 1325.250000 C 159.000000 1362.000000 159.000000 1405.199951 159.000000 1433.250000 C 159.000000 1461.300049 159.000000 1498.699951 159.000000 1526.750000 C 159.000000 1554.800049 159.000000 1601.199951 159.000000 1642.750000 C 159.000000 1684.300049 159.000000 1739.699951 159.000000 1781.250000 C 159.000000 1822.800049 159.000000 1870.800049 159.000000 1901.250000 C 159.000000 1931.699951 159.000000 1972.300049 159.000000 2002.750000 C 159.000000 2033.199951 159.000000 2073.800049 159.000000 2104.250000 C 159.000000 2134.699951 159.000000 2175.300049 159.000000 2205.750000 C 159.000000 2236.199951 159.000000 2281.000000 159.000000 2317.750000 C 159.000000 2354.500000 159.000000 2403.500000 159.000000 2440.250000 C 159.000000 2477.000000 159.000000 2526.000000 159.000000 2562.750000 C 159.000000 2599.500000 159.000000 2758.699951 159.000000 2803.500000" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="159.000000" y="1731.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">Yes</text></g><g class="KEhhc0RlY2xhcmVkQ29uZmlnIC0mZ3Q7IENoZWNrUmVnaXN0ZXJlZClbMF0="><path d="M 204.350138 628.475509 C 259.000000 688.200012 290.799988 721.400024 359.179506 790.163662" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="281.000000" y="717.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">No</text></g><g class="KENoZWNrUmVnaXN0ZXJlZCAtJmd0OyBGb3JFYWNoUmVnaXN0ZXJlZClbMF0="><path d="M 413.980001 921.999900 C 413.600006 960.000000 413.500000 980.000000 413.500000 1016.000000" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /></g><g class="KEZvckVhY2hSZWdpc3RlcmVkIC0mZ3Q7IEhhc1JlZ2lzdGVyZWRDb25maWcpWzBd"><path d="M 413.500000 1104.000000 C 413.500000 1142.000000 413.600006 1161.800049 413.959186 1197.000208" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /></g><g class="KEhhc1JlZ2lzdGVyZWRDb25maWcgLSZndDsgTmV4dFJlZ2lzdGVyZWQpWzBd"><path d="M 345.265308 1301.995411 C 227.800003 1369.400024 198.000000 1405.199951 198.000000 1433.250000 C 198.000000 1461.300049 198.000000 1498.699951 198.000000 1526.750000 C 198.000000 1554.800049 198.000000 1601.199951 198.000000 1642.750000 C 198.000000 1684.300049 303.299988 1867.697021 720.552299 1935.840270" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="242.000000" y="1766.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">No</text></g><g class="KEhhc1JlZ2lzdGVyZWRDb25maWcgLSZndDsgTG9hZFJlZ2lzdGVyZWRDb25maWcpWzBd"><path d="M 447.096053 1315.672922 C 484.000000 1372.000000 493.500000 1398.699951 493.500000 1443.500000" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="485.000000" y="1381.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">Yes</text></g><g class="KExvYWRSZWdpc3RlcmVkQ29uZmlnIC0mZ3Q7IENoZWNrQ29udGV4dClbMF0="><path d="M 493.500000 1514.500000 C 493.500000 1561.300049 493.600006 1585.599976 493.966944 1630.000137" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /></g><g class="KENoZWNrQ29udGV4dCAtJmd0OyBMb2FkUmVnaXN0ZXJlZClbMF0="><path d="M 451.876233 1772.654433 C 409.799988 1834.599976 399.000000 1870.800049 399.000000 1901.250000 C 399.000000 1931.699951 399.000000 1972.300049 399.000000 2002.750000 C 399.000000 2033.199951 399.000000 2073.800049 399.000000 2104.250000 C 399.000000 2134.699951 399.000000 2175.300049 399.000000 2205.750000 C 399.000000 2236.199951 399.000000 2281.000000 399.000000 2317.750000 C 399.000000 2354.500000 399.000000 2403.500000 399.000000 2440.250000 C 399.000000 2477.000000 399.000000 2526.000000 399.000000 2562.750000 C 399.000000 2599.500000 399.000000 2758.699951 399.000000 2803.500000" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="399.000000" y="2287.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">Yes</text></g><g class="KENoZWNrQ29udGV4dCAtJmd0OyBOZXh0UmVnaXN0ZXJlZClbMF0="><path d="M 576.789759 1752.892615 C 733.000000 1830.800049 778.099976 1862.699951 798.831329 1907.864682" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="703.000000" y="1821.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">No</text></g><g class="KE5leHRSZWdpc3RlcmVkIC0mZ3Q7IEhhc1JlZ2lzdGVyZWRDb25maWcpWzBd"><path d="M 825.823550 1909.526345 C 833.500000 1862.699951 835.500000 1822.800049 835.500000 1781.250000 C 835.500000 1739.699951 835.500000 1684.300049 835.500000 1642.750000 C 835.500000 1601.199951 835.500000 1554.800049 835.500000 1526.750000 C 835.500000 1498.699951 769.400024 1367.400024 508.842788 1292.110397" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="834.000000" y="1500.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">More notebooks</text></g><g class="KE5leHRSZWdpc3RlcmVkIC0mZ3Q7IFN0YXJ0QW5jZXN0b3JTZWFyY2gpWzBd"><path d="M 819.500000 1994.500000 C 819.500000 2041.300049 819.500000 2065.698975 819.500000 2110.500000" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="820.000000" y="2059.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">No more notebooks</text></g><g class="KFN0YXJ0QW5jZXN0b3JTZWFyY2ggLSZndDsgSXNSb290KVswXQ=="><path d="M 819.500000 2197.500000 C 819.500000 2244.300049 819.599976 2268.399902 819.966386 2312.000141" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /></g><g class="KElzUm9vdCAtJmd0OyBOb3RGb3VuZClbMF0="><path d="M 748.149482 2407.758673 C 565.599976 2482.600098 519.599976 2517.399902 519.974842 2577.000079" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="602.000000" y="2473.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">Yes</text></g><g class="KElzUm9vdCAtJmd0OyBIYXNBbmNlc3RvckNvbmZpZylbMF0="><path d="M 820.033053 2442.999727 C 820.799988 2489.399902 821.000000 2513.399902 821.000000 2557.000000" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="821.000000" y="2506.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">No</text></g><g class="KEhhc0FuY2VzdG9yQ29uZmlnIC0mZ3Q7IExvYWRBbmNlc3RvcilbMF0="><path d="M 782.815848 2674.611764 C 740.799988 2731.800049 730.000000 2758.699951 730.000000 2803.500000" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="740.000000" y="2740.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">Yes</text></g><g class="KEhhc0FuY2VzdG9yQ29uZmlnIC0mZ3Q7IEdvVG9QYXJlbnQpWzBd"><path d="M 866.368200 2671.458776 C 922.400024 2731.199951 948.750000 2758.699951 993.945046 2804.648293" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /><text x="930.000000" y="2745.000000" fill="#676C7E" class="text-italic fill-N2" style="text-anchor:middle;font-size:16px">No</text></g><g class="KEdvVG9QYXJlbnQgLSZndDsgSXNSb290KVswXQ=="><path d="M 1043.041920 2805.521419 C 1049.949951 2758.699951 1051.750000 2722.000000 1051.750000 2685.250000 C 1051.750000 2648.500000 1017.799988 2483.399902 885.529701 2412.881811" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /></g><g class="KExvYWREZWNsYXJlZCAtJmd0OyBTdWNjZXNzKVswXQ=="><path d="M 159.000000 2891.000000 C 159.000000 2929.000000 184.199997 2951.000000 281.388558 2997.280266" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /></g><g class="KExvYWRSZWdpc3RlcmVkIC0mZ3Q7IFN1Y2Nlc3MpWzBd"><path d="M 399.000000 2891.000000 C 399.000000 2929.000000 397.000000 2949.000000 389.784465 2985.077677" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /></g><g class="KExvYWRBbmNlc3RvciAtJmd0OyBTdWNjZXNzKVswXQ=="><path d="M 730.000000 2891.000000 C 730.000000 2929.000000 684.000000 2952.600098 503.835865 3005.865920" stroke="#0D32B2" fill="none" class="connection stroke-B1" style="stroke-width:2;" marker-end="url(#mk-d2-3094856511-3488378134)" mask="url(#d2-3094856511)" /></g><mask id="d2-3094856511" maskUnits="userSpaceOnUse" x="-101" y="-101" width="1320" height="3300">
<rect x="-101" y="-101" width="1320" height="3300" fill="white"></rect>
<rect x="128.000000" y="427.000000" width="97" height="21" fill="black"></rect>
<rect x="400.000000" y="548.000000" width="56" height="21" fill="black"></rect>
<rect x="146.000000" y="1715.000000" width="26" height="21" fill="black"></rect>
<rect x="270.000000" y="701.000000" width="22" height="21" fill="black"></rect>
<rect x="231.000000" y="1750.000000" width="22" height="21" fill="black"></rect>
<rect x="472.000000" y="1365.000000" width="26" height="21" fill="black"></rect>
<rect x="386.000000" y="2271.000000" width="26" height="21" fill="black"></rect>
<rect x="692.000000" y="1805.000000" width="22" height="21" fill="black"></rect>
<rect x="779.000000" y="1484.000000" width="110" height="21" fill="black"></rect>
<rect x="754.000000" y="2043.000000" width="132" height="21" fill="black"></rect>
<rect x="589.000000" y="2457.000000" width="26" height="21" fill="black"></rect>
<rect x="810.000000" y="2490.000000" width="22" height="21" fill="black"></rect>
<rect x="727.000000" y="2724.000000" width="26" height="21" fill="black"></rect>
<rect x="919.000000" y="2729.000000" width="22" height="21" fill="black"></rect>
</mask></svg></svg>

### 1. Declared Path (Tier 1 - Highest Priority)

The system first checks if a notebook path has been explicitly declared via:
- CLI flag: `opennotes --notebook /path/to/notebook`
- Environment variable: `OPENNOTES_NOTEBOOK=/path/to/notebook`

If a declared path exists:
1. Check if `.opennotes.json` exists in that path
2. If yes: Load and open the notebook → **SUCCESS**
3. If no: Continue to Tier 2

### 2. Registered Notebooks (Tier 2 - Context Matching)

The system checks notebooks registered in the global configuration:

1. Load global config from `~/.config/opennotes/config.json`
2. For each registered notebook path:
   - Check if `.opennotes.json` exists
   - If exists: Load notebook configuration
   - Check if current directory matches any context path using **Context Matching Algorithm**
   - If match found: Load and open the notebook → **SUCCESS**
3. If no registered notebooks match: Continue to Tier 3

#### Context Matching Algorithm

```go
// For each context path in notebook config
for _, context := range notebook.Contexts {
    if strings.HasPrefix(currentWorkingDirectory, context) {
        return notebook // Match found
    }
}
```

**Example:**
```
Notebook contexts: ["/home/user/project", "/tmp/work"]
Current directory: "/home/user/project/src"

Match check: strings.HasPrefix("/home/user/project/src", "/home/user/project")
Result: TRUE → Context matches → Return this notebook
```

### 3. Ancestor Search (Tier 3 - Fallback)

If no declared or registered notebooks match, the system performs an ancestor directory search:

1. Start with current working directory
2. Check if `.opennotes.json` exists in current directory
3. If yes: Load and open the notebook → **SUCCESS**
4. If no: Move to parent directory
5. Repeat until reaching filesystem root (`/`) or empty string
6. If root reached: → **NOT FOUND**

## File Locations & Formats

### Global Configuration
**Location:** `~/.config/opennotes/config.json`

```json
{
  "notebooks": [
    "/home/user/work-notebook",
    "/home/user/personal-notebook",
    "/tmp/temp-notebook"
  ]
}
```

### Notebook Configuration
**Location:** `<notebook_directory>/.opennotes.json`

```json
{
  "root": "./notes",
  "name": "Project Notebook",
  "contexts": [
    "/home/user/project",
    "/home/user/project-related"
  ],
  "templates": {
    "default": "# {{.Title}}\n\nDate: {{.Date}}\n\n"
  },
  "groups": [
    {
      "name": "Default",
      "globs": ["**/*.md"],
      "metadata": {}
    }
  ]
}
```

## Key Features

### Deterministic Behavior
- **Clear Priority**: Declared > Registered > Ancestor
- **First Match Wins**: Stops at first successful discovery
- **No Ambiguity**: Priority order prevents conflicts

### Graceful Fallback
- If higher priority method fails, try next tier
- Comprehensive search ensures maximum discovery success
- Returns `nil` only when all methods exhausted

### Context-Aware
- Registered notebooks define active contexts
- Automatically selects appropriate notebook for current work environment
- Supports multiple context paths per notebook

### Efficient Discovery
- Stops immediately upon successful match
- Minimal filesystem operations
- Fast context matching using string prefix comparison

## State Transitions Summary

1. **DECLARED PATH** → Success or Continue to Tier 2
2. **REGISTERED SEARCH** → For each registered notebook:
   - Check exists → Check context match → Success or Continue
3. **ANCESTOR SEARCH** → Walk up directories until found or root
4. **SUCCESS** → Return notebook instance
5. **NOT FOUND** → Return nil

This discovery system ensures OpenNotes works seamlessly across different project environments while maintaining predictable, efficient behavior.

