# Claude tokenizer calibration

Generated: 2026-07-23T03:01:09Z

Method: Claude framing tokens = count_tokens("x") - 1 known probe token; Claude content tokens = count_tokens(message text) - framing tokens; factor = sum(Claude content tokens) / sum(o200k_base tokens); held-out samples are excluded from fitting and evaluated only after factors are finalized

| Generation | Model | Framing baseline | Samples | o200k tokens | Claude content tokens | Factor | MAPE |
|---|---|---:|---:|---:|---:|---:|---:|
| claude | `claude-opus-4-7` | 11 | 80 | 17684 | 29256 | 1.654377 | 8.13% |
| claude-legacy | `claude-opus-4-6` | 7 | 80 | 17684 | 22109 | 1.250226 | 5.16% |
| fable5 | `claude-fable-5` | 6 | 80 | 17684 | 29256 | 1.654377 | 8.13% |

## claude (`claude-opus-4-7`)

| Language | Samples | o200k tokens | Claude content tokens | Fitted factor | Mean ratio | Median ratio | Fitted-factor MAPE | Global-factor signed aggregate error | Global-factor MAPE | LOO language MAPE |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| C# | 5 | 890 | 1698 | 1.907865 | 1.914918 | 1.879121 | 5.85% | -13.31% | 13.25% | 7.43% |
| C++ | 5 | 1244 | 1938 | 1.557878 | 1.547032 | 1.512195 | 5.08% | +6.19% | 7.23% | 6.38% |
| Go | 5 | 1073 | 1767 | 1.646785 | 1.634693 | 1.672340 | 5.87% | +0.51% | 5.80% | 7.27% |
| JSON | 5 | 829 | 1228 | 1.481303 | 1.473328 | 1.450000 | 4.04% | +11.73% | 12.52% | 4.90% |
| Java | 5 | 1026 | 1810 | 1.764133 | 1.775699 | 1.803468 | 5.91% | -6.19% | 7.94% | 7.41% |
| JavaScript | 5 | 1183 | 1944 | 1.643280 | 1.655184 | 1.636735 | 4.83% | +0.67% | 4.93% | 5.99% |
| Markdown | 5 | 882 | 1406 | 1.594104 | 1.600131 | 1.614035 | 6.51% | +3.77% | 6.96% | 8.13% |
| PHP | 5 | 1419 | 2157 | 1.520085 | 1.518787 | 1.551839 | 4.56% | +8.85% | 9.30% | 5.99% |
| Python | 5 | 1151 | 1925 | 1.672459 | 1.674680 | 1.647619 | 2.72% | -1.09% | 2.45% | 3.40% |
| Ruby | 5 | 994 | 1538 | 1.547284 | 1.551221 | 1.521951 | 3.34% | +6.89% | 6.90% | 4.23% |
| Rust | 5 | 1624 | 2713 | 1.670567 | 1.665562 | 1.691710 | 6.25% | -1.00% | 6.34% | 7.75% |
| SQL | 5 | 1394 | 2785 | 1.997848 | 2.026809 | 1.914201 | 8.35% | -17.16% | 17.22% | 10.52% |
| Shell | 5 | 1457 | 2403 | 1.649279 | 1.658476 | 1.705085 | 4.43% | +0.37% | 4.41% | 5.57% |
| TOML | 5 | 637 | 1007 | 1.580848 | 1.576610 | 1.586207 | 3.05% | +4.67% | 5.10% | 3.79% |
| TypeScript | 5 | 1150 | 1870 | 1.626087 | 1.632146 | 1.609195 | 6.07% | +1.76% | 6.58% | 7.67% |
| YAML | 5 | 731 | 1067 | 1.459644 | 1.460944 | 1.473333 | 1.50% | +13.31% | 13.24% | 1.94% |

### Held-out evaluation

These samples were not used to fit or select calibration factors. Factor source: production calibration factors.

| Language | Samples | Factor used | Factor basis | o200k tokens | Claude content tokens | Predicted tokens | Signed aggregate error | MAPE |
|---|---:|---:|---|---:|---:|---:|---:|---:|
| C | 5 | 1.654377 | global fallback | 1412 | 2240 | 2336 | +4.29% | 7.05% |
| HTML | 5 | 1.654377 | global fallback | 1710 | 2413 | 2829 | +17.24% | 17.27% |
| Kotlin | 5 | 1.654377 | global fallback | 1249 | 2148 | 2067 | -3.77% | 5.16% |
| Swift | 5 | 1.654377 | global fallback | 1260 | 2083 | 2085 | +0.10% | 3.51% |
| **Overall** | **20** |  |  | **5631** | **8884** | **9317** | **+4.87%** | **8.25%** |

<details>
<summary>Held-out per-file measurements</summary>

| File | Language | Bytes | Truncated | o200k | Claude content | Ratio |
|---|---|---:|:---:|---:|---:|---:|
| `holdout/c/csv_parser.c` | C | 1171 | false | 267 | 409 | 1.531835 |
| `holdout/c/metrics.c` | C | 1058 | false | 260 | 389 | 1.496154 |
| `holdout/c/path_table.c` | C | 1128 | false | 280 | 491 | 1.753571 |
| `holdout/c/ring_buffer.c` | C | 993 | false | 289 | 441 | 1.525952 |
| `holdout/c/scheduler.c` | C | 1301 | false | 316 | 510 | 1.613924 |
| `holdout/html/checkout.html` | HTML | 1396 | false | 372 | 534 | 1.435484 |
| `holdout/html/dashboard.html` | HTML | 1243 | false | 402 | 552 | 1.373134 |
| `holdout/html/docs.html` | HTML | 1034 | false | 318 | 457 | 1.437107 |
| `holdout/html/settings.html` | HTML | 1289 | false | 319 | 462 | 1.448276 |
| `holdout/html/status.html` | HTML | 1038 | false | 299 | 408 | 1.364548 |
| `holdout/kotlin/BatchWindow.kt` | Kotlin | 1080 | false | 253 | 443 | 1.750988 |
| `holdout/kotlin/ConfigMerge.kt` | Kotlin | 1146 | false | 252 | 430 | 1.706349 |
| `holdout/kotlin/EventRouter.kt` | Kotlin | 1110 | false | 227 | 427 | 1.881057 |
| `holdout/kotlin/RetryQueue.kt` | Kotlin | 1023 | false | 251 | 423 | 1.685259 |
| `holdout/kotlin/Stats.kt` | Kotlin | 1039 | false | 266 | 425 | 1.597744 |
| `holdout/swift/AsyncCache.swift` | Swift | 1022 | false | 248 | 400 | 1.612903 |
| `holdout/swift/EventBus.swift` | Swift | 1012 | false | 241 | 380 | 1.576763 |
| `holdout/swift/RetryPolicy.swift` | Swift | 1179 | false | 273 | 489 | 1.791209 |
| `holdout/swift/RouteMatcher.swift` | Swift | 1201 | false | 249 | 405 | 1.626506 |
| `holdout/swift/SlidingWindow.swift` | Swift | 1065 | false | 249 | 409 | 1.642570 |

</details>

<details>
<summary>Per-file measurements</summary>

| File | Language | Bytes | Truncated | o200k | Claude content | Ratio |
|---|---|---:|:---:|---:|---:|---:|
| `corpus/csharp/AsyncBatcher.cs` | C# | 950 | false | 182 | 342 | 1.879121 |
| `corpus/csharp/Inventory.cs` | C# | 697 | false | 135 | 282 | 2.088889 |
| `corpus/csharp/MetricsWindow.cs` | C# | 936 | false | 192 | 334 | 1.739583 |
| `corpus/csharp/Result.cs` | C# | 845 | false | 173 | 318 | 1.838150 |
| `corpus/csharp/RouteMatcher.cs` | C# | 1048 | false | 208 | 422 | 2.028846 |
| `corpus/cpp/bounded_queue.cpp` | C++ | 1309 | false | 306 | 506 | 1.653595 |
| `corpus/cpp/config_parser.cpp` | C++ | 1023 | false | 246 | 372 | 1.512195 |
| `corpus/cpp/graph.cpp` | C++ | 1196 | false | 269 | 439 | 1.631970 |
| `corpus/cpp/record_formatter.cpp` | C++ | 938 | false | 223 | 325 | 1.457399 |
| `corpus/cpp/window.cpp` | C++ | 815 | false | 200 | 296 | 1.480000 |
| `corpus/go/cache.go` | Go | 784 | false | 227 | 400 | 1.762115 |
| `corpus/go/ledger.go` | Go | 665 | false | 175 | 281 | 1.605714 |
| `corpus/go/parser.go` | Go | 914 | false | 254 | 433 | 1.704724 |
| `corpus/go/stream.go` | Go | 659 | false | 182 | 260 | 1.428571 |
| `corpus/go/worker.go` | Go | 901 | false | 235 | 393 | 1.672340 |
| `corpus/config/catalog.json` | JSON | 594 | false | 192 | 277 | 1.442708 |
| `corpus/config/deployment.json` | JSON | 587 | false | 189 | 295 | 1.560847 |
| `corpus/config/feature_flags.json` | JSON | 579 | false | 180 | 261 | 1.450000 |
| `corpus/config/policy.json` | JSON | 628 | false | 167 | 255 | 1.526946 |
| `corpus/config/service.json` | JSON | 309 | false | 101 | 140 | 1.386139 |
| `corpus/java/AsyncMemoizer.java` | Java | 1049 | false | 205 | 391 | 1.907317 |
| `corpus/java/CsvRow.java` | Java | 1149 | false | 229 | 365 | 1.593886 |
| `corpus/java/EventRouter.java` | Java | 858 | false | 173 | 312 | 1.803468 |
| `corpus/java/RetryPolicy.java` | Java | 824 | false | 169 | 316 | 1.869822 |
| `corpus/java/SlidingStats.java` | Java | 1174 | false | 250 | 426 | 1.704000 |
| `corpus/javascript/cache.js` | JavaScript | 602 | false | 149 | 263 | 1.765101 |
| `corpus/javascript/metrics.js` | JavaScript | 975 | false | 275 | 418 | 1.520000 |
| `corpus/javascript/queue.js` | JavaScript | 1058 | false | 257 | 412 | 1.603113 |
| `corpus/javascript/retry.js` | JavaScript | 1013 | false | 245 | 401 | 1.636735 |
| `corpus/javascript/settings.js` | JavaScript | 1106 | false | 257 | 450 | 1.750973 |
| `corpus/markdown/api.md` | Markdown | 757 | false | 196 | 278 | 1.418367 |
| `corpus/markdown/architecture.md` | Markdown | 887 | false | 172 | 294 | 1.709302 |
| `corpus/markdown/contributing.md` | Markdown | 911 | false | 171 | 276 | 1.614035 |
| `corpus/markdown/migration.md` | Markdown | 815 | false | 176 | 269 | 1.528409 |
| `corpus/markdown/runbook.md` | Markdown | 922 | false | 167 | 289 | 1.730539 |
| `corpus/php/InvoiceService.php` | PHP | 1483 | false | 332 | 534 | 1.608434 |
| `corpus/php/middleware.php` | PHP | 1370 | false | 310 | 486 | 1.567742 |
| `corpus/php/paginator.php` | PHP | 1334 | false | 333 | 456 | 1.369369 |
| `corpus/php/routes.php` | PHP | 624 | false | 145 | 217 | 1.496552 |
| `corpus/php/webhook.php` | PHP | 1360 | false | 299 | 464 | 1.551839 |
| `corpus/python/batching.py` | Python | 913 | false | 210 | 346 | 1.647619 |
| `corpus/python/config.py` | Python | 1034 | false | 244 | 416 | 1.704918 |
| `corpus/python/events.py` | Python | 948 | false | 231 | 376 | 1.627706 |
| `corpus/python/inventory.py` | Python | 915 | false | 205 | 361 | 1.760976 |
| `corpus/python/retry.py` | Python | 1138 | false | 261 | 426 | 1.632184 |
| `corpus/ruby/circuit_breaker.rb` | Ruby | 732 | false | 208 | 310 | 1.490385 |
| `corpus/ruby/dependency_graph.rb` | Ruby | 781 | false | 190 | 315 | 1.657895 |
| `corpus/ruby/invoice.rb` | Ruby | 860 | false | 227 | 342 | 1.506608 |
| `corpus/ruby/json_lines.rb` | Ruby | 816 | false | 205 | 312 | 1.521951 |
| `corpus/ruby/slug.rb` | Ruby | 585 | false | 164 | 259 | 1.579268 |
| `corpus/rust/catalog.rs` | Rust | 1498 | false | 378 | 671 | 1.775132 |
| `corpus/rust/parser.rs` | Rust | 1139 | false | 285 | 451 | 1.582456 |
| `corpus/rust/retry.rs` | Rust | 1175 | false | 272 | 486 | 1.786765 |
| `corpus/rust/statistics.rs` | Rust | 1175 | false | 303 | 452 | 1.491749 |
| `corpus/rust/worker_pool.rs` | Rust | 1500 | false | 386 | 653 | 1.691710 |
| `corpus/sql/cohort_report.sql` | SQL | 1427 | false | 338 | 647 | 1.914201 |
| `corpus/sql/inventory_schema.sql` | SQL | 1063 | false | 238 | 600 | 2.521008 |
| `corpus/sql/reconciliation.sql` | SQL | 1464 | false | 333 | 610 | 1.831832 |
| `corpus/sql/retention.sql` | SQL | 761 | false | 175 | 351 | 2.005714 |
| `corpus/sql/session_rollup.sql` | SQL | 1309 | false | 310 | 577 | 1.861290 |
| `corpus/shell/backup.sh` | Shell | 545 | false | 172 | 298 | 1.732558 |
| `corpus/shell/deploy.sh` | Shell | 1072 | false | 295 | 503 | 1.705085 |
| `corpus/shell/healthcheck.sh` | Shell | 1012 | false | 328 | 502 | 1.530488 |
| `corpus/shell/import_csv.sh` | Shell | 1010 | false | 328 | 563 | 1.716463 |
| `corpus/shell/prune_artifacts.sh` | Shell | 972 | false | 334 | 537 | 1.607784 |
| `corpus/config/database.toml` | TOML | 497 | false | 135 | 220 | 1.629630 |
| `corpus/config/formatter.toml` | TOML | 486 | false | 137 | 224 | 1.635036 |
| `corpus/config/logging.toml` | TOML | 504 | false | 140 | 219 | 1.564286 |
| `corpus/config/service.toml` | TOML | 348 | false | 109 | 160 | 1.467890 |
| `corpus/config/workspace.toml` | TOML | 447 | false | 116 | 184 | 1.586207 |
| `corpus/typescript/client.ts` | TypeScript | 758 | false | 188 | 297 | 1.579787 |
| `corpus/typescript/graph.ts` | TypeScript | 994 | false | 261 | 420 | 1.609195 |
| `corpus/typescript/registry.ts` | TypeScript | 949 | false | 215 | 363 | 1.688372 |
| `corpus/typescript/result.ts` | TypeScript | 991 | false | 264 | 384 | 1.454545 |
| `corpus/typescript/scheduler.ts` | TypeScript | 916 | false | 222 | 406 | 1.828829 |
| `corpus/config/access.yaml` | YAML | 577 | false | 150 | 221 | 1.473333 |
| `corpus/config/jobs.yaml` | YAML | 505 | false | 156 | 231 | 1.480769 |
| `corpus/config/observability.yml` | YAML | 531 | false | 161 | 229 | 1.422360 |
| `corpus/config/pipeline.yaml` | YAML | 408 | false | 118 | 175 | 1.483051 |
| `corpus/config/release.yaml` | YAML | 526 | false | 146 | 211 | 1.445205 |

</details>

## claude-legacy (`claude-opus-4-6`)

| Language | Samples | o200k tokens | Claude content tokens | Fitted factor | Mean ratio | Median ratio | Fitted-factor MAPE | Global-factor signed aggregate error | Global-factor MAPE | LOO language MAPE |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| C# | 5 | 890 | 1250 | 1.404494 | 1.408521 | 1.428571 | 4.26% | -10.96% | 11.02% | 5.34% |
| C++ | 5 | 1244 | 1594 | 1.281350 | 1.281618 | 1.278027 | 0.60% | -2.38% | 2.40% | 0.82% |
| Go | 5 | 1073 | 1405 | 1.309413 | 1.304691 | 1.295276 | 3.69% | -4.41% | 4.58% | 4.64% |
| JSON | 5 | 829 | 986 | 1.189385 | 1.189685 | 1.188119 | 0.56% | +5.07% | 5.04% | 0.75% |
| Java | 5 | 1026 | 1373 | 1.338207 | 1.342237 | 1.319527 | 4.57% | -6.63% | 6.60% | 5.87% |
| JavaScript | 5 | 1183 | 1494 | 1.262891 | 1.276626 | 1.249027 | 4.40% | -1.07% | 4.20% | 5.36% |
| Markdown | 5 | 882 | 987 | 1.119048 | 1.119015 | 1.122449 | 1.35% | +11.75% | 11.79% | 1.67% |
| PHP | 5 | 1419 | 1723 | 1.214235 | 1.212301 | 1.203226 | 1.97% | +2.96% | 3.84% | 2.56% |
| Python | 5 | 1151 | 1420 | 1.233710 | 1.235333 | 1.242424 | 1.72% | +1.34% | 1.71% | 2.21% |
| Ruby | 5 | 994 | 1173 | 1.180080 | 1.182298 | 1.207317 | 3.65% | +5.97% | 5.98% | 4.61% |
| Rust | 5 | 1624 | 2092 | 1.288177 | 1.287118 | 1.298246 | 1.85% | -2.92% | 3.54% | 2.29% |
| SQL | 5 | 1394 | 1782 | 1.278336 | 1.285559 | 1.302521 | 3.12% | -2.13% | 3.84% | 3.94% |
| Shell | 5 | 1457 | 1744 | 1.196980 | 1.198772 | 1.196610 | 1.81% | +4.47% | 4.36% | 2.30% |
| TOML | 5 | 637 | 776 | 1.218210 | 1.217625 | 1.220183 | 1.74% | +2.58% | 2.66% | 2.24% |
| TypeScript | 5 | 1150 | 1452 | 1.262609 | 1.264539 | 1.291188 | 4.16% | -0.96% | 4.30% | 5.30% |
| YAML | 5 | 731 | 858 | 1.173735 | 1.173751 | 1.173913 | 0.67% | +6.64% | 6.67% | 1.04% |

### Held-out evaluation

These samples were not used to fit or select calibration factors. Factor source: production calibration factors.

| Language | Samples | Factor used | Factor basis | o200k tokens | Claude content tokens | Predicted tokens | Signed aggregate error | MAPE |
|---|---:|---:|---|---:|---:|---:|---:|---:|
| C | 5 | 1.250226 | global fallback | 1412 | 1783 | 1765 | -1.01% | 1.24% |
| HTML | 5 | 1.250226 | global fallback | 1710 | 1995 | 2139 | +7.22% | 7.26% |
| Kotlin | 5 | 1.250226 | global fallback | 1249 | 1560 | 1562 | +0.13% | 3.36% |
| Swift | 5 | 1.250226 | global fallback | 1260 | 1551 | 1574 | +1.48% | 4.25% |
| **Overall** | **20** |  |  | **5631** | **6889** | **7040** | **+2.19%** | **4.03%** |

<details>
<summary>Held-out per-file measurements</summary>

| File | Language | Bytes | Truncated | o200k | Claude content | Ratio |
|---|---|---:|:---:|---:|---:|---:|
| `holdout/c/csv_parser.c` | C | 1171 | false | 267 | 344 | 1.288390 |
| `holdout/c/metrics.c` | C | 1058 | false | 260 | 324 | 1.246154 |
| `holdout/c/path_table.c` | C | 1128 | false | 280 | 356 | 1.271429 |
| `holdout/c/ring_buffer.c` | C | 993 | false | 289 | 360 | 1.245675 |
| `holdout/c/scheduler.c` | C | 1301 | false | 316 | 399 | 1.262658 |
| `holdout/html/checkout.html` | HTML | 1396 | false | 372 | 436 | 1.172043 |
| `holdout/html/dashboard.html` | HTML | 1243 | false | 402 | 469 | 1.166667 |
| `holdout/html/docs.html` | HTML | 1034 | false | 318 | 373 | 1.172956 |
| `holdout/html/settings.html` | HTML | 1289 | false | 319 | 373 | 1.169279 |
| `holdout/html/status.html` | HTML | 1038 | false | 299 | 344 | 1.150502 |
| `holdout/kotlin/BatchWindow.kt` | Kotlin | 1080 | false | 253 | 313 | 1.237154 |
| `holdout/kotlin/ConfigMerge.kt` | Kotlin | 1146 | false | 252 | 304 | 1.206349 |
| `holdout/kotlin/EventRouter.kt` | Kotlin | 1110 | false | 227 | 302 | 1.330396 |
| `holdout/kotlin/RetryQueue.kt` | Kotlin | 1023 | false | 251 | 321 | 1.278884 |
| `holdout/kotlin/Stats.kt` | Kotlin | 1039 | false | 266 | 320 | 1.203008 |
| `holdout/swift/AsyncCache.swift` | Swift | 1022 | false | 248 | 293 | 1.181452 |
| `holdout/swift/EventBus.swift` | Swift | 1012 | false | 241 | 291 | 1.207469 |
| `holdout/swift/RetryPolicy.swift` | Swift | 1179 | false | 273 | 363 | 1.329670 |
| `holdout/swift/RouteMatcher.swift` | Swift | 1201 | false | 249 | 301 | 1.208835 |
| `holdout/swift/SlidingWindow.swift` | Swift | 1065 | false | 249 | 303 | 1.216867 |

</details>

<details>
<summary>Per-file measurements</summary>

| File | Language | Bytes | Truncated | o200k | Claude content | Ratio |
|---|---|---:|:---:|---:|---:|---:|
| `corpus/csharp/AsyncBatcher.cs` | C# | 950 | false | 182 | 260 | 1.428571 |
| `corpus/csharp/Inventory.cs` | C# | 697 | false | 135 | 203 | 1.503704 |
| `corpus/csharp/MetricsWindow.cs` | C# | 936 | false | 192 | 257 | 1.338542 |
| `corpus/csharp/Result.cs` | C# | 845 | false | 173 | 230 | 1.329480 |
| `corpus/csharp/RouteMatcher.cs` | C# | 1048 | false | 208 | 300 | 1.442308 |
| `corpus/cpp/bounded_queue.cpp` | C++ | 1309 | false | 306 | 389 | 1.271242 |
| `corpus/cpp/config_parser.cpp` | C++ | 1023 | false | 246 | 314 | 1.276423 |
| `corpus/cpp/graph.cpp` | C++ | 1196 | false | 269 | 349 | 1.297398 |
| `corpus/cpp/record_formatter.cpp` | C++ | 938 | false | 223 | 285 | 1.278027 |
| `corpus/cpp/window.cpp` | C++ | 815 | false | 200 | 257 | 1.285000 |
| `corpus/go/cache.go` | Go | 784 | false | 227 | 320 | 1.409692 |
| `corpus/go/ledger.go` | Go | 665 | false | 175 | 222 | 1.268571 |
| `corpus/go/parser.go` | Go | 914 | false | 254 | 329 | 1.295276 |
| `corpus/go/stream.go` | Go | 659 | false | 182 | 224 | 1.230769 |
| `corpus/go/worker.go` | Go | 901 | false | 235 | 310 | 1.319149 |
| `corpus/config/catalog.json` | JSON | 594 | false | 192 | 226 | 1.177083 |
| `corpus/config/deployment.json` | JSON | 587 | false | 189 | 224 | 1.185185 |
| `corpus/config/feature_flags.json` | JSON | 579 | false | 180 | 215 | 1.194444 |
| `corpus/config/policy.json` | JSON | 628 | false | 167 | 201 | 1.203593 |
| `corpus/config/service.json` | JSON | 309 | false | 101 | 120 | 1.188119 |
| `corpus/java/AsyncMemoizer.java` | Java | 1049 | false | 205 | 306 | 1.492683 |
| `corpus/java/CsvRow.java` | Java | 1149 | false | 229 | 290 | 1.266376 |
| `corpus/java/EventRouter.java` | Java | 858 | false | 173 | 234 | 1.352601 |
| `corpus/java/RetryPolicy.java` | Java | 824 | false | 169 | 223 | 1.319527 |
| `corpus/java/SlidingStats.java` | Java | 1174 | false | 250 | 320 | 1.280000 |
| `corpus/javascript/cache.js` | JavaScript | 602 | false | 149 | 212 | 1.422819 |
| `corpus/javascript/metrics.js` | JavaScript | 975 | false | 275 | 333 | 1.210909 |
| `corpus/javascript/queue.js` | JavaScript | 1058 | false | 257 | 330 | 1.284047 |
| `corpus/javascript/retry.js` | JavaScript | 1013 | false | 245 | 298 | 1.216327 |
| `corpus/javascript/settings.js` | JavaScript | 1106 | false | 257 | 321 | 1.249027 |
| `corpus/markdown/api.md` | Markdown | 757 | false | 196 | 220 | 1.122449 |
| `corpus/markdown/architecture.md` | Markdown | 887 | false | 172 | 196 | 1.139535 |
| `corpus/markdown/contributing.md` | Markdown | 911 | false | 171 | 186 | 1.087719 |
| `corpus/markdown/migration.md` | Markdown | 815 | false | 176 | 196 | 1.113636 |
| `corpus/markdown/runbook.md` | Markdown | 922 | false | 167 | 189 | 1.131737 |
| `corpus/php/InvoiceService.php` | PHP | 1483 | false | 332 | 422 | 1.271084 |
| `corpus/php/middleware.php` | PHP | 1370 | false | 310 | 373 | 1.203226 |
| `corpus/php/paginator.php` | PHP | 1334 | false | 333 | 394 | 1.183183 |
| `corpus/php/routes.php` | PHP | 624 | false | 145 | 174 | 1.200000 |
| `corpus/php/webhook.php` | PHP | 1360 | false | 299 | 360 | 1.204013 |
| `corpus/python/batching.py` | Python | 913 | false | 210 | 265 | 1.261905 |
| `corpus/python/config.py` | Python | 1034 | false | 244 | 291 | 1.192623 |
| `corpus/python/events.py` | Python | 948 | false | 231 | 287 | 1.242424 |
| `corpus/python/inventory.py` | Python | 915 | false | 205 | 257 | 1.253659 |
| `corpus/python/retry.py` | Python | 1138 | false | 261 | 320 | 1.226054 |
| `corpus/ruby/circuit_breaker.rb` | Ruby | 732 | false | 208 | 226 | 1.086538 |
| `corpus/ruby/dependency_graph.rb` | Ruby | 781 | false | 190 | 233 | 1.226316 |
| `corpus/ruby/invoice.rb` | Ruby | 860 | false | 227 | 266 | 1.171806 |
| `corpus/ruby/json_lines.rb` | Ruby | 816 | false | 205 | 250 | 1.219512 |
| `corpus/ruby/slug.rb` | Ruby | 585 | false | 164 | 198 | 1.207317 |
| `corpus/rust/catalog.rs` | Rust | 1498 | false | 378 | 488 | 1.291005 |
| `corpus/rust/parser.rs` | Rust | 1139 | false | 285 | 370 | 1.298246 |
| `corpus/rust/retry.rs` | Rust | 1175 | false | 272 | 355 | 1.305147 |
| `corpus/rust/statistics.rs` | Rust | 1175 | false | 303 | 372 | 1.227723 |
| `corpus/rust/worker_pool.rs` | Rust | 1500 | false | 386 | 507 | 1.313472 |
| `corpus/sql/cohort_report.sql` | SQL | 1427 | false | 338 | 442 | 1.307692 |
| `corpus/sql/inventory_schema.sql` | SQL | 1063 | false | 238 | 310 | 1.302521 |
| `corpus/sql/reconciliation.sql` | SQL | 1464 | false | 333 | 403 | 1.210210 |
| `corpus/sql/retention.sql` | SQL | 761 | false | 175 | 235 | 1.342857 |
| `corpus/sql/session_rollup.sql` | SQL | 1309 | false | 310 | 392 | 1.264516 |
| `corpus/shell/backup.sh` | Shell | 545 | false | 172 | 209 | 1.215116 |
| `corpus/shell/deploy.sh` | Shell | 1072 | false | 295 | 353 | 1.196610 |
| `corpus/shell/healthcheck.sh` | Shell | 1012 | false | 328 | 383 | 1.167683 |
| `corpus/shell/import_csv.sh` | Shell | 1010 | false | 328 | 406 | 1.237805 |
| `corpus/shell/prune_artifacts.sh` | Shell | 972 | false | 334 | 393 | 1.176647 |
| `corpus/config/database.toml` | TOML | 497 | false | 135 | 169 | 1.251852 |
| `corpus/config/formatter.toml` | TOML | 486 | false | 137 | 169 | 1.233577 |
| `corpus/config/logging.toml` | TOML | 504 | false | 140 | 167 | 1.192857 |
| `corpus/config/service.toml` | TOML | 348 | false | 109 | 133 | 1.220183 |
| `corpus/config/workspace.toml` | TOML | 447 | false | 116 | 138 | 1.189655 |
| `corpus/typescript/client.ts` | TypeScript | 758 | false | 188 | 231 | 1.228723 |
| `corpus/typescript/graph.ts` | TypeScript | 994 | false | 261 | 337 | 1.291188 |
| `corpus/typescript/registry.ts` | TypeScript | 949 | false | 215 | 288 | 1.339535 |
| `corpus/typescript/result.ts` | TypeScript | 991 | false | 264 | 309 | 1.170455 |
| `corpus/typescript/scheduler.ts` | TypeScript | 916 | false | 222 | 287 | 1.292793 |
| `corpus/config/access.yaml` | YAML | 577 | false | 150 | 173 | 1.153333 |
| `corpus/config/jobs.yaml` | YAML | 505 | false | 156 | 186 | 1.192308 |
| `corpus/config/observability.yml` | YAML | 531 | false | 161 | 189 | 1.173913 |
| `corpus/config/pipeline.yaml` | YAML | 408 | false | 118 | 139 | 1.177966 |
| `corpus/config/release.yaml` | YAML | 526 | false | 146 | 171 | 1.171233 |

</details>

## fable5 (`claude-fable-5`)

| Language | Samples | o200k tokens | Claude content tokens | Fitted factor | Mean ratio | Median ratio | Fitted-factor MAPE | Global-factor signed aggregate error | Global-factor MAPE | LOO language MAPE |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| C# | 5 | 890 | 1698 | 1.907865 | 1.914918 | 1.879121 | 5.85% | -13.31% | 13.25% | 7.43% |
| C++ | 5 | 1244 | 1938 | 1.557878 | 1.547032 | 1.512195 | 5.08% | +6.19% | 7.23% | 6.38% |
| Go | 5 | 1073 | 1767 | 1.646785 | 1.634693 | 1.672340 | 5.87% | +0.51% | 5.80% | 7.27% |
| JSON | 5 | 829 | 1228 | 1.481303 | 1.473328 | 1.450000 | 4.04% | +11.73% | 12.52% | 4.90% |
| Java | 5 | 1026 | 1810 | 1.764133 | 1.775699 | 1.803468 | 5.91% | -6.19% | 7.94% | 7.41% |
| JavaScript | 5 | 1183 | 1944 | 1.643280 | 1.655184 | 1.636735 | 4.83% | +0.67% | 4.93% | 5.99% |
| Markdown | 5 | 882 | 1406 | 1.594104 | 1.600131 | 1.614035 | 6.51% | +3.77% | 6.96% | 8.13% |
| PHP | 5 | 1419 | 2157 | 1.520085 | 1.518787 | 1.551839 | 4.56% | +8.85% | 9.30% | 5.99% |
| Python | 5 | 1151 | 1925 | 1.672459 | 1.674680 | 1.647619 | 2.72% | -1.09% | 2.45% | 3.40% |
| Ruby | 5 | 994 | 1538 | 1.547284 | 1.551221 | 1.521951 | 3.34% | +6.89% | 6.90% | 4.23% |
| Rust | 5 | 1624 | 2713 | 1.670567 | 1.665562 | 1.691710 | 6.25% | -1.00% | 6.34% | 7.75% |
| SQL | 5 | 1394 | 2785 | 1.997848 | 2.026809 | 1.914201 | 8.35% | -17.16% | 17.22% | 10.52% |
| Shell | 5 | 1457 | 2403 | 1.649279 | 1.658476 | 1.705085 | 4.43% | +0.37% | 4.41% | 5.57% |
| TOML | 5 | 637 | 1007 | 1.580848 | 1.576610 | 1.586207 | 3.05% | +4.67% | 5.10% | 3.79% |
| TypeScript | 5 | 1150 | 1870 | 1.626087 | 1.632146 | 1.609195 | 6.07% | +1.76% | 6.58% | 7.67% |
| YAML | 5 | 731 | 1067 | 1.459644 | 1.460944 | 1.473333 | 1.50% | +13.31% | 13.24% | 1.94% |

### Held-out evaluation

These samples were not used to fit or select calibration factors. Factor source: training global factor (no production mapping).

| Language | Samples | Factor used | Factor basis | o200k tokens | Claude content tokens | Predicted tokens | Signed aggregate error | MAPE |
|---|---:|---:|---|---:|---:|---:|---:|---:|
| C | 5 | 1.654377 | global fallback | 1412 | 2240 | 2336 | +4.29% | 7.05% |
| HTML | 5 | 1.654377 | global fallback | 1710 | 2413 | 2829 | +17.24% | 17.27% |
| Kotlin | 5 | 1.654377 | global fallback | 1249 | 2148 | 2067 | -3.77% | 5.16% |
| Swift | 5 | 1.654377 | global fallback | 1260 | 2083 | 2085 | +0.10% | 3.51% |
| **Overall** | **20** |  |  | **5631** | **8884** | **9317** | **+4.87%** | **8.25%** |

<details>
<summary>Held-out per-file measurements</summary>

| File | Language | Bytes | Truncated | o200k | Claude content | Ratio |
|---|---|---:|:---:|---:|---:|---:|
| `holdout/c/csv_parser.c` | C | 1171 | false | 267 | 409 | 1.531835 |
| `holdout/c/metrics.c` | C | 1058 | false | 260 | 389 | 1.496154 |
| `holdout/c/path_table.c` | C | 1128 | false | 280 | 491 | 1.753571 |
| `holdout/c/ring_buffer.c` | C | 993 | false | 289 | 441 | 1.525952 |
| `holdout/c/scheduler.c` | C | 1301 | false | 316 | 510 | 1.613924 |
| `holdout/html/checkout.html` | HTML | 1396 | false | 372 | 534 | 1.435484 |
| `holdout/html/dashboard.html` | HTML | 1243 | false | 402 | 552 | 1.373134 |
| `holdout/html/docs.html` | HTML | 1034 | false | 318 | 457 | 1.437107 |
| `holdout/html/settings.html` | HTML | 1289 | false | 319 | 462 | 1.448276 |
| `holdout/html/status.html` | HTML | 1038 | false | 299 | 408 | 1.364548 |
| `holdout/kotlin/BatchWindow.kt` | Kotlin | 1080 | false | 253 | 443 | 1.750988 |
| `holdout/kotlin/ConfigMerge.kt` | Kotlin | 1146 | false | 252 | 430 | 1.706349 |
| `holdout/kotlin/EventRouter.kt` | Kotlin | 1110 | false | 227 | 427 | 1.881057 |
| `holdout/kotlin/RetryQueue.kt` | Kotlin | 1023 | false | 251 | 423 | 1.685259 |
| `holdout/kotlin/Stats.kt` | Kotlin | 1039 | false | 266 | 425 | 1.597744 |
| `holdout/swift/AsyncCache.swift` | Swift | 1022 | false | 248 | 400 | 1.612903 |
| `holdout/swift/EventBus.swift` | Swift | 1012 | false | 241 | 380 | 1.576763 |
| `holdout/swift/RetryPolicy.swift` | Swift | 1179 | false | 273 | 489 | 1.791209 |
| `holdout/swift/RouteMatcher.swift` | Swift | 1201 | false | 249 | 405 | 1.626506 |
| `holdout/swift/SlidingWindow.swift` | Swift | 1065 | false | 249 | 409 | 1.642570 |

</details>

<details>
<summary>Per-file measurements</summary>

| File | Language | Bytes | Truncated | o200k | Claude content | Ratio |
|---|---|---:|:---:|---:|---:|---:|
| `corpus/csharp/AsyncBatcher.cs` | C# | 950 | false | 182 | 342 | 1.879121 |
| `corpus/csharp/Inventory.cs` | C# | 697 | false | 135 | 282 | 2.088889 |
| `corpus/csharp/MetricsWindow.cs` | C# | 936 | false | 192 | 334 | 1.739583 |
| `corpus/csharp/Result.cs` | C# | 845 | false | 173 | 318 | 1.838150 |
| `corpus/csharp/RouteMatcher.cs` | C# | 1048 | false | 208 | 422 | 2.028846 |
| `corpus/cpp/bounded_queue.cpp` | C++ | 1309 | false | 306 | 506 | 1.653595 |
| `corpus/cpp/config_parser.cpp` | C++ | 1023 | false | 246 | 372 | 1.512195 |
| `corpus/cpp/graph.cpp` | C++ | 1196 | false | 269 | 439 | 1.631970 |
| `corpus/cpp/record_formatter.cpp` | C++ | 938 | false | 223 | 325 | 1.457399 |
| `corpus/cpp/window.cpp` | C++ | 815 | false | 200 | 296 | 1.480000 |
| `corpus/go/cache.go` | Go | 784 | false | 227 | 400 | 1.762115 |
| `corpus/go/ledger.go` | Go | 665 | false | 175 | 281 | 1.605714 |
| `corpus/go/parser.go` | Go | 914 | false | 254 | 433 | 1.704724 |
| `corpus/go/stream.go` | Go | 659 | false | 182 | 260 | 1.428571 |
| `corpus/go/worker.go` | Go | 901 | false | 235 | 393 | 1.672340 |
| `corpus/config/catalog.json` | JSON | 594 | false | 192 | 277 | 1.442708 |
| `corpus/config/deployment.json` | JSON | 587 | false | 189 | 295 | 1.560847 |
| `corpus/config/feature_flags.json` | JSON | 579 | false | 180 | 261 | 1.450000 |
| `corpus/config/policy.json` | JSON | 628 | false | 167 | 255 | 1.526946 |
| `corpus/config/service.json` | JSON | 309 | false | 101 | 140 | 1.386139 |
| `corpus/java/AsyncMemoizer.java` | Java | 1049 | false | 205 | 391 | 1.907317 |
| `corpus/java/CsvRow.java` | Java | 1149 | false | 229 | 365 | 1.593886 |
| `corpus/java/EventRouter.java` | Java | 858 | false | 173 | 312 | 1.803468 |
| `corpus/java/RetryPolicy.java` | Java | 824 | false | 169 | 316 | 1.869822 |
| `corpus/java/SlidingStats.java` | Java | 1174 | false | 250 | 426 | 1.704000 |
| `corpus/javascript/cache.js` | JavaScript | 602 | false | 149 | 263 | 1.765101 |
| `corpus/javascript/metrics.js` | JavaScript | 975 | false | 275 | 418 | 1.520000 |
| `corpus/javascript/queue.js` | JavaScript | 1058 | false | 257 | 412 | 1.603113 |
| `corpus/javascript/retry.js` | JavaScript | 1013 | false | 245 | 401 | 1.636735 |
| `corpus/javascript/settings.js` | JavaScript | 1106 | false | 257 | 450 | 1.750973 |
| `corpus/markdown/api.md` | Markdown | 757 | false | 196 | 278 | 1.418367 |
| `corpus/markdown/architecture.md` | Markdown | 887 | false | 172 | 294 | 1.709302 |
| `corpus/markdown/contributing.md` | Markdown | 911 | false | 171 | 276 | 1.614035 |
| `corpus/markdown/migration.md` | Markdown | 815 | false | 176 | 269 | 1.528409 |
| `corpus/markdown/runbook.md` | Markdown | 922 | false | 167 | 289 | 1.730539 |
| `corpus/php/InvoiceService.php` | PHP | 1483 | false | 332 | 534 | 1.608434 |
| `corpus/php/middleware.php` | PHP | 1370 | false | 310 | 486 | 1.567742 |
| `corpus/php/paginator.php` | PHP | 1334 | false | 333 | 456 | 1.369369 |
| `corpus/php/routes.php` | PHP | 624 | false | 145 | 217 | 1.496552 |
| `corpus/php/webhook.php` | PHP | 1360 | false | 299 | 464 | 1.551839 |
| `corpus/python/batching.py` | Python | 913 | false | 210 | 346 | 1.647619 |
| `corpus/python/config.py` | Python | 1034 | false | 244 | 416 | 1.704918 |
| `corpus/python/events.py` | Python | 948 | false | 231 | 376 | 1.627706 |
| `corpus/python/inventory.py` | Python | 915 | false | 205 | 361 | 1.760976 |
| `corpus/python/retry.py` | Python | 1138 | false | 261 | 426 | 1.632184 |
| `corpus/ruby/circuit_breaker.rb` | Ruby | 732 | false | 208 | 310 | 1.490385 |
| `corpus/ruby/dependency_graph.rb` | Ruby | 781 | false | 190 | 315 | 1.657895 |
| `corpus/ruby/invoice.rb` | Ruby | 860 | false | 227 | 342 | 1.506608 |
| `corpus/ruby/json_lines.rb` | Ruby | 816 | false | 205 | 312 | 1.521951 |
| `corpus/ruby/slug.rb` | Ruby | 585 | false | 164 | 259 | 1.579268 |
| `corpus/rust/catalog.rs` | Rust | 1498 | false | 378 | 671 | 1.775132 |
| `corpus/rust/parser.rs` | Rust | 1139 | false | 285 | 451 | 1.582456 |
| `corpus/rust/retry.rs` | Rust | 1175 | false | 272 | 486 | 1.786765 |
| `corpus/rust/statistics.rs` | Rust | 1175 | false | 303 | 452 | 1.491749 |
| `corpus/rust/worker_pool.rs` | Rust | 1500 | false | 386 | 653 | 1.691710 |
| `corpus/sql/cohort_report.sql` | SQL | 1427 | false | 338 | 647 | 1.914201 |
| `corpus/sql/inventory_schema.sql` | SQL | 1063 | false | 238 | 600 | 2.521008 |
| `corpus/sql/reconciliation.sql` | SQL | 1464 | false | 333 | 610 | 1.831832 |
| `corpus/sql/retention.sql` | SQL | 761 | false | 175 | 351 | 2.005714 |
| `corpus/sql/session_rollup.sql` | SQL | 1309 | false | 310 | 577 | 1.861290 |
| `corpus/shell/backup.sh` | Shell | 545 | false | 172 | 298 | 1.732558 |
| `corpus/shell/deploy.sh` | Shell | 1072 | false | 295 | 503 | 1.705085 |
| `corpus/shell/healthcheck.sh` | Shell | 1012 | false | 328 | 502 | 1.530488 |
| `corpus/shell/import_csv.sh` | Shell | 1010 | false | 328 | 563 | 1.716463 |
| `corpus/shell/prune_artifacts.sh` | Shell | 972 | false | 334 | 537 | 1.607784 |
| `corpus/config/database.toml` | TOML | 497 | false | 135 | 220 | 1.629630 |
| `corpus/config/formatter.toml` | TOML | 486 | false | 137 | 224 | 1.635036 |
| `corpus/config/logging.toml` | TOML | 504 | false | 140 | 219 | 1.564286 |
| `corpus/config/service.toml` | TOML | 348 | false | 109 | 160 | 1.467890 |
| `corpus/config/workspace.toml` | TOML | 447 | false | 116 | 184 | 1.586207 |
| `corpus/typescript/client.ts` | TypeScript | 758 | false | 188 | 297 | 1.579787 |
| `corpus/typescript/graph.ts` | TypeScript | 994 | false | 261 | 420 | 1.609195 |
| `corpus/typescript/registry.ts` | TypeScript | 949 | false | 215 | 363 | 1.688372 |
| `corpus/typescript/result.ts` | TypeScript | 991 | false | 264 | 384 | 1.454545 |
| `corpus/typescript/scheduler.ts` | TypeScript | 916 | false | 222 | 406 | 1.828829 |
| `corpus/config/access.yaml` | YAML | 577 | false | 150 | 221 | 1.473333 |
| `corpus/config/jobs.yaml` | YAML | 505 | false | 156 | 231 | 1.480769 |
| `corpus/config/observability.yml` | YAML | 531 | false | 161 | 229 | 1.422360 |
| `corpus/config/pipeline.yaml` | YAML | 408 | false | 118 | 175 | 1.483051 |
| `corpus/config/release.yaml` | YAML | 526 | false | 146 | 211 | 1.445205 |

</details>
