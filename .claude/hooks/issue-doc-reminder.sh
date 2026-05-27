#!/bin/bash
# PreToolUse フック（Bash 用）。
#
# 目的: Issue をクローズする操作（`gh issue close` / `gh pr merge`）の前に、
# 対応する Issue へ「取り組み内容・改善・結果」のコメントを投稿し忘れない
# ようにする注意喚起。
#
# 仕組み・限界:
#   - フックはコメントが実際に投稿されたかを検証できない。そこで「確認済み」を
#     表す目印 ISSUE_DOCUMENTED がコマンドに含まれていれば許可、無ければ deny
#     して理由を返す（Claude はコメント投稿後、コマンド末尾に
#     ` # ISSUE_DOCUMENTED` を付けて再実行する）。
#   - PR マージ時の "Closes #N" は GitHub サーバー側で自動クローズされるため、
#     ローカルのフックでは捕捉できない（CLAUDE.md ルールで補う）。

input=$(cat)
cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // ""')

# 対象は Issue をクローズし得る操作のみ。
if printf '%s' "$cmd" | grep -Eq 'gh +issue +close|gh +pr +merge'; then
  if printf '%s' "$cmd" | grep -q 'ISSUE_DOCUMENTED'; then
    exit 0 # 確認済み → 許可
  fi
  reason='Issue をクローズ/PR をマージする前に、対応する Issue へ取り組み内容（実装・改善・結果・残課題）のコメントを投稿してください: gh issue comment <番号> -b "..."。投稿済み、または Issue を閉じない操作であれば、コマンド末尾に " # ISSUE_DOCUMENTED" を付けて再実行してください。'
  jq -n --arg r "$reason" '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: $r
    }
  }'
  exit 0
fi

exit 0
