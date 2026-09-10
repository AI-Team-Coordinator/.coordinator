#!/bin/bash
# Sync Common/data/progress with the orphan branch origin/coordinator-state.
# Common HEAD stays on main. Agents keep reading data/progress/ on disk.
#
# Usage: coordinator_state.sh {ensure|pull|push} [commit-message]

set -e
export GIT_TERMINAL_PROMPT=0

# shellcheck source=paths.sh
. "$(dirname "$0")/paths.sh"

STATE_BRANCH="${COORDINATOR_STATE_BRANCH:-coordinator-state}"
STATE_WORKTREE="${COORDINATOR_STATE_WORKTREE:-$COMMON_ROOT/.coordinator-state}"
RSYNC_EXCLUDES=(--exclude '.sync.log' --exclude '*.lock' --exclude '.DS_Store')

usage() {
    echo "Usage: $0 {ensure|pull|push} [commit-message]" >&2
    exit 1
}

assert_worktree_path() {
    local common_abs wt_abs
    common_abs=$(cd "$COMMON_ROOT" && pwd)
    wt_abs=$(cd "$STATE_WORKTREE" 2>/dev/null && pwd || true)
    if [ -z "$wt_abs" ]; then
        return 0
    fi
    if [ "$wt_abs" = "$common_abs" ]; then
        echo "coordinator_state: refusing to use Common root as worktree" >&2
        exit 1
    fi
}

is_linked_worktree() {
    [ -f "$STATE_WORKTREE/.git" ]
}

install_hooks() {
    if [ -d "$COMMON_ROOT/.githooks" ]; then
        git -C "$COMMON_ROOT" config core.hooksPath .githooks
    fi
}

copy_progress() {
    local src=$1 dst=$2
    shift 2
    mkdir -p "$dst"
    rsync -a "$@" "${RSYNC_EXCLUDES[@]}" "$src/" "$dst/"
}

write_state_gitignore() {
    cat > "$STATE_WORKTREE/.gitignore" <<'EOF'
.DS_Store
data/progress/.sync.log
data/progress/**/*.lock
EOF
}

seed_orphan() {
    assert_worktree_path
    if [ -e "$STATE_WORKTREE" ] && [ ! -f "$STATE_WORKTREE/.git" ]; then
        echo "coordinator_state: $STATE_WORKTREE exists and is not a git worktree" >&2
        exit 1
    fi
    if [ ! -e "$STATE_WORKTREE" ]; then
        git -C "$COMMON_ROOT" worktree add --detach "$STATE_WORKTREE"
    fi
    if [ ! -f "$STATE_WORKTREE/.git" ]; then
        echo "coordinator_state: expected linked worktree at $STATE_WORKTREE" >&2
        exit 1
    fi
    git -C "$STATE_WORKTREE" checkout --orphan "$STATE_BRANCH"
    git -C "$STATE_WORKTREE" rm -rf --ignore-unmatch . >/dev/null 2>&1 || true
    find "$STATE_WORKTREE" -mindepth 1 -maxdepth 1 ! -name '.git' -exec rm -rf {} +
    mkdir -p "$STATE_WORKTREE/data/progress"
    write_state_gitignore
    if [ -d "$PROGRESS_DIR" ]; then
        copy_progress "$PROGRESS_DIR" "$STATE_WORKTREE/data/progress"
    fi
    git -C "$STATE_WORKTREE" add -A
    git -C "$STATE_WORKTREE" commit --allow-empty -m "chore(progress): seed $STATE_BRANCH" >/dev/null
}

ensure() {
    install_hooks
    mkdir -p "$PROGRESS_DIR"
    assert_worktree_path

    if is_linked_worktree; then
        local head
        head=$(git -C "$STATE_WORKTREE" rev-parse --abbrev-ref HEAD 2>/dev/null || true)
        if [ "$head" = "$STATE_BRANCH" ]; then
            return 0
        fi
    fi

    git -C "$COMMON_ROOT" fetch origin "refs/heads/${STATE_BRANCH}:refs/remotes/origin/${STATE_BRANCH}" >/dev/null 2>&1 || true

    if is_linked_worktree; then
        if git -C "$COMMON_ROOT" show-ref --verify --quiet "refs/heads/${STATE_BRANCH}"; then
            git -C "$STATE_WORKTREE" checkout -B "$STATE_BRANCH" "$STATE_BRANCH" >/dev/null
            return 0
        fi
        if git -C "$COMMON_ROOT" show-ref --verify --quiet "refs/remotes/origin/${STATE_BRANCH}"; then
            git -C "$STATE_WORKTREE" checkout -B "$STATE_BRANCH" "origin/${STATE_BRANCH}" >/dev/null
            return 0
        fi
        seed_orphan
        return 0
    fi

    if git -C "$COMMON_ROOT" show-ref --verify --quiet "refs/heads/${STATE_BRANCH}"; then
        git -C "$COMMON_ROOT" worktree add "$STATE_WORKTREE" "$STATE_BRANCH"
        return 0
    fi
    if git -C "$COMMON_ROOT" show-ref --verify --quiet "refs/remotes/origin/${STATE_BRANCH}"; then
        git -C "$COMMON_ROOT" worktree add -b "$STATE_BRANCH" "$STATE_WORKTREE" "origin/${STATE_BRANCH}"
        return 0
    fi

    seed_orphan
}

pull_state() {
    ensure
    git -C "$COMMON_ROOT" fetch origin "$STATE_BRANCH" >/dev/null 2>&1 || true
    if git -C "$COMMON_ROOT" show-ref --verify --quiet "refs/remotes/origin/${STATE_BRANCH}"; then
        git -C "$STATE_WORKTREE" merge --ff-only "origin/${STATE_BRANCH}" >/dev/null 2>&1 || \
            git -C "$STATE_WORKTREE" pull --rebase origin "$STATE_BRANCH" >/dev/null 2>&1 || true
    fi
    if [ -d "$STATE_WORKTREE/data/progress" ]; then
        copy_progress "$STATE_WORKTREE/data/progress" "$PROGRESS_DIR" --update
    fi
}

push_state() {
    local msg=${1:-chore(progress): update}
    ensure
    mkdir -p "$PROGRESS_DIR" "$STATE_WORKTREE/data/progress"
    write_state_gitignore
    copy_progress "$PROGRESS_DIR" "$STATE_WORKTREE/data/progress"
    git -C "$STATE_WORKTREE" add -A
    git -C "$STATE_WORKTREE" commit -m "$msg" >/dev/null 2>&1 || true
    git -C "$STATE_WORKTREE" pull --rebase origin "$STATE_BRANCH" >/dev/null 2>&1 || true
    if [ -d "$STATE_WORKTREE/data/progress" ]; then
        copy_progress "$STATE_WORKTREE/data/progress" "$PROGRESS_DIR" --update
    fi
    git -C "$STATE_WORKTREE" push -u origin "$STATE_BRANCH" >/dev/null 2>&1 || true
}

CMD=${1:-}
shift || true

case "$CMD" in
    ensure) ensure ;;
    pull) pull_state ;;
    push) push_state "$*" ;;
    *) usage ;;
esac
