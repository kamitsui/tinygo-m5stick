# ============================================================
#  tinygo-m5stick — Makefile
#  TinyGo を使った M5StickC Plus2 開発のビルド/書き込みを集約する。
#
#  前提: tinygo / esptool が PATH 上にあること（Homebrew 導入を想定。
#        Apple Silicon では /opt/homebrew/bin が PATH に必要）。
#
#  使い方の例:
#    make list                       # projects/ の一覧
#    make build PROJ=01-blink        # ビルド
#    make flash PROJ=02-display      # 実機へ書き込み（ポートは自動検出）
#    make flash PROJ=03-buzzer PORT=/dev/cu.xxx   # ポート指定
# ============================================================

# --- 設定（環境変数 / コマンドラインで上書き可） ---
TARGET ?= esp32-coreboard-v2
PROJ   ?= 01-blink
# シリアルポート自動検出（M5StickC Plus2 は WCH USB シリアル）。
PORT   ?= $(firstword $(wildcard /dev/cu.wchusbserial*))
TINYGO ?= tinygo

PROJ_DIR  := projects/$(PROJ)
BUILD_DIR := build
OUT       := $(BUILD_DIR)/$(PROJ).bin

.DEFAULT_GOAL := help
.PHONY: help list check-tools check-proj build flash monitor clean docs-dev docs-build docs-preview

help: ## このヘルプを表示
	@echo "tinygo-m5stick — targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "vars: PROJ=$(PROJ)  TARGET=$(TARGET)  PORT=$(if $(PORT),$(PORT),<auto: 未検出>)"

list: ## projects/ のプロジェクト一覧を表示
	@ls -1 projects 2>/dev/null || echo "(projects/ が見つかりません)"

check-tools: ## tinygo が PATH 上にあるか確認
	@command -v $(TINYGO) >/dev/null 2>&1 || { \
		echo "ERROR: '$(TINYGO)' が見つかりません。Homebrew で導入し PATH を通してください:"; \
		echo "       brew tap tinygo-org/tools && brew install tinygo esptool"; \
		exit 1; }

check-proj:
	@test -d "$(PROJ_DIR)" || { echo "ERROR: $(PROJ_DIR) が存在しません。'make list' で確認してください"; exit 1; }

build: check-tools check-proj ## ビルド (例: make build PROJ=01-blink)
	@mkdir -p $(BUILD_DIR)
	$(TINYGO) build -target=$(TARGET) -o $(OUT) ./$(PROJ_DIR)
	@echo "built: $(OUT)"

flash: check-tools check-proj ## 実機へ書き込み (例: make flash PROJ=02-display [PORT=...])
	@test -n "$(PORT)" || { \
		echo "ERROR: シリアルポートが見つかりません。デバイスを接続するか PORT=/dev/cu.xxx を指定してください"; \
		echo "       接続中のポート: $$(ls /dev/cu.* 2>/dev/null | tr '\n' ' ')"; \
		exit 1; }
	$(TINYGO) flash -target=$(TARGET) -port=$(PORT) ./$(PROJ_DIR)

monitor: check-tools ## シリアルモニタを開く (例: make monitor [PORT=...])
	$(TINYGO) monitor -port=$(PORT)

clean: ## ビルド生成物を削除
	rm -rf $(BUILD_DIR)

# ---- ドキュメント (VitePress) ----
docs-dev: ## ドキュメントをローカルプレビュー（開発サーバ）
	npm run docs:dev

docs-build: ## ドキュメントをビルド
	npm run docs:build

docs-preview: ## ビルド済みドキュメントをプレビュー
	npm run docs:preview
