#!/usr/bin/env sh
set -eu

env_value() {
  key="$1"
  if [ -f .env ]; then
    sed -n "s/^$key=//p" .env | tail -n 1
  fi
}

BASE_URL="${BASE_URL:-http://localhost:8000}"
SMOKE_USERNAME="${SMOKE_USERNAME:-smoke_$(date +%s)}"
SMOKE_PASSWORD="${SMOKE_PASSWORD:-password123}"
SUPER_ADMIN_USERNAME="${SUPER_ADMIN_USERNAME:-$(env_value SUPER_ADMIN_USERNAME)}"
SUPER_ADMIN_PASSWORD="${SUPER_ADMIN_PASSWORD:-$(env_value SUPER_ADMIN_PASSWORD)}"
SUPER_ADMIN_USERNAME="${SUPER_ADMIN_USERNAME:-admin}"
SUPER_ADMIN_PASSWORD="${SUPER_ADMIN_PASSWORD:-admin_password_change_me}"
TMP_DIR="${TMPDIR:-/tmp}/novel-reader-smoke.$$"

TOKEN=""
SUPER_ADMIN_TOKEN=""
ADMIN_TOKEN=""
CATEGORY_ID=""
SMOKE_USER_ID=""
SMOKE_BOOK_ID=""

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

need_curl() {
  if ! command -v curl >/dev/null 2>&1; then
    echo "curl is required for smoke checks" >&2
    exit 127
  fi
}

request() {
  method="$1"
  path="$2"
  body="${3:-}"
  auth="${4:-}"
  output="$TMP_DIR/response.json"
  headers="$TMP_DIR/headers.txt"
  : > "$output"
  : > "$headers"

  if [ -n "$body" ]; then
    if [ -n "$auth" ]; then
      code="$(curl -sS -X "$method" "$BASE_URL$path" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $auth" \
        -d "$body" \
        -D "$headers" \
        -o "$output" \
        -w "%{http_code}")" || code="000"
    else
      code="$(curl -sS -X "$method" "$BASE_URL$path" \
        -H "Content-Type: application/json" \
        -d "$body" \
        -D "$headers" \
        -o "$output" \
        -w "%{http_code}")" || code="000"
    fi
  else
    if [ -n "$auth" ]; then
      code="$(curl -sS -X "$method" "$BASE_URL$path" \
        -H "Authorization: Bearer $auth" \
        -D "$headers" \
        -o "$output" \
        -w "%{http_code}")" || code="000"
    else
      code="$(curl -sS -X "$method" "$BASE_URL$path" \
        -D "$headers" \
        -o "$output" \
        -w "%{http_code}")" || code="000"
    fi
  fi

  printf '%s' "$code"
}

assert_status() {
  name="$1"
  code="$2"
  expected="$3"

  case ",$expected," in
    *,"$code",*)
      echo "PASS $name ($code)"
      ;;
    *)
      echo "FAIL $name: expected $expected, got $code" >&2
      echo "Response:" >&2
      sed -n '1,120p' "$TMP_DIR/response.json" >&2 || true
      exit 1
      ;;
  esac
}

json_value() {
  key="$1"
  sed -n "s/.*\"$key\"[[:space:]]*:[[:space:]]*\"\\([^\"]*\\)\".*/\\1/p" "$TMP_DIR/response.json" | head -n 1
}

json_number() {
  key="$1"
  sed -n "s/.*\"$key\"[[:space:]]*:[[:space:]]*\\([0-9][0-9]*\\).*/\\1/p" "$TMP_DIR/response.json" | head -n 1
}

check_health() {
  code="$(request GET /health)"
  case "$code" in
    200)
      echo "PASS backend health /health (200)"
      ;;
    404|405)
      code="$(request GET /api/health)"
      assert_status "backend health /api/health" "$code" "200"
      ;;
    *)
      assert_status "backend health" "$code" "200"
      ;;
  esac
}

register_and_login() {
  body="{\"username\":\"$SMOKE_USERNAME\",\"password\":\"$SMOKE_PASSWORD\"}"

  code="$(request POST /api/auth/register "$body")"
  assert_status "register smoke user" "$code" "200,201,409"

  code="$(request POST /api/auth/login "$body")"
  assert_status "login smoke user" "$code" "200"
  TOKEN="$(json_value token)"

  if [ -z "$TOKEN" ]; then
    echo "FAIL login smoke user: response does not contain token" >&2
    sed -n '1,120p' "$TMP_DIR/response.json" >&2 || true
    exit 1
  fi

  code="$(request GET /api/auth/me "" "$TOKEN")"
  assert_status "current user" "$code" "200"
  SMOKE_USER_ID="$(json_number id)"
  if [ -z "$SMOKE_USER_ID" ]; then
    echo "FAIL current user: response does not contain id" >&2
    sed -n '1,120p' "$TMP_DIR/response.json" >&2 || true
    exit 1
  fi
}

check_public_books() {
  code="$(request GET /api/categories)"
  assert_status "categories" "$code" "200"
  CATEGORY_ID="$(json_number id)"
  if [ -z "$CATEGORY_ID" ]; then
    echo "FAIL categories: response does not contain category id" >&2
    sed -n '1,120p' "$TMP_DIR/response.json" >&2 || true
    exit 1
  fi

  code="$(request GET "/api/books/search?q=&page=1&pageSize=10")"
  assert_status "book search" "$code" "200"

  code="$(request GET "/api/books/search?q=no_such_novel_$SMOKE_USERNAME&page=1&pageSize=10")"
  assert_status "empty book search" "$code" "200"
}

check_account_management() {
  code="$(request PATCH /api/auth/me "{\"nickname\":\"Smoke Nick\"}" "$TOKEN")"
  assert_status "update nickname" "$code" "200"

  new_password="${SMOKE_PASSWORD}_new"
  body="{\"oldPassword\":\"$SMOKE_PASSWORD\",\"newPassword\":\"$new_password\"}"
  code="$(request PATCH /api/auth/password "$body" "$TOKEN")"
  assert_status "change password" "$code" "200"

  login_body="{\"username\":\"$SMOKE_USERNAME\",\"password\":\"$new_password\"}"
  code="$(request POST /api/auth/login "$login_body")"
  assert_status "login with changed password" "$code" "200"
  TOKEN="$(json_value token)"
  SMOKE_PASSWORD="$new_password"
}

relogin_smoke_user() {
  login_body="{\"username\":\"$SMOKE_USERNAME\",\"password\":\"$SMOKE_PASSWORD\"}"
  code="$(request POST /api/auth/login "$login_body")"
  assert_status "relogin smoke user" "$code" "200"
  TOKEN="$(json_value token)"
  if [ -z "$TOKEN" ]; then
    echo "FAIL relogin smoke user: response does not contain token" >&2
    sed -n '1,120p' "$TMP_DIR/response.json" >&2 || true
    exit 1
  fi

  code="$(request GET /api/auth/me "" "$TOKEN")"
  assert_status "current user after relogin" "$code" "200"
}

check_avatar_upload() {
  sample="$TMP_DIR/avatar-test.jpeg"
  cp data/uploads/covers/base.jpeg "$sample"

  code="$(curl -sS -X POST "$BASE_URL/api/auth/me/avatar" \
    -H "Authorization: Bearer $TOKEN" \
    -F "file=@$sample;type=image/jpeg" \
    -o "$TMP_DIR/response.json" \
    -w "%{http_code}")" || code="000"

  assert_status "avatar upload" "$code" "200"
  code="$(request GET /api/auth/me "" "$TOKEN")"
  assert_status "current user after avatar upload" "$code" "200"
  avatar_url="$(json_value avatarUrl)"
  if [ -z "$avatar_url" ]; then
    echo "FAIL avatar upload: response does not contain avatarUrl" >&2
    sed -n '1,120p' "$TMP_DIR/response.json" >&2 || true
    exit 1
  fi

  code="$(curl -sS -D "$TMP_DIR/avatar-headers.txt" -o "$TMP_DIR/avatar-body.bin" -w "%{http_code}" "$BASE_URL$avatar_url")" || code="000"
  assert_status "avatar fetch" "$code" "200"
  if ! grep -qi '^content-type: image/jpeg' "$TMP_DIR/avatar-headers.txt" && ! grep -qi '^content-type: image/jpg' "$TMP_DIR/avatar-headers.txt"; then
    echo "FAIL avatar fetch: expected image/jpeg content type" >&2
    sed -n '1,40p' "$TMP_DIR/avatar-headers.txt" >&2 || true
    exit 1
  fi
}

check_bookshelf_management() {
  if [ -z "$SMOKE_BOOK_ID" ]; then
    echo "FAIL bookshelf smoke: missing book id" >&2
    exit 1
  fi

  code="$(request GET /api/me/bookshelf/groups "" "$TOKEN")"
  assert_status "bookshelf groups" "$code" "200"

  code="$(request POST "/api/me/bookshelf/$SMOKE_BOOK_ID" "{}" "$TOKEN")"
  assert_status "add book to bookshelf" "$code" "200"

  code="$(request GET "/api/me/bookshelf?page=1&pageSize=10" "" "$TOKEN")"
  assert_status "root bookshelf list" "$code" "200"

  code="$(request POST /api/me/bookshelf/groups "{\"name\":\"Smoke Shelf\"}" "$TOKEN")"
  assert_status "create bookshelf group" "$code" "200"
  custom_group_id="$(json_number id)"
  if [ -z "$custom_group_id" ]; then
    echo "FAIL create bookshelf group: response does not contain id" >&2
    sed -n '1,120p' "$TMP_DIR/response.json" >&2 || true
    exit 1
  fi

  code="$(request PATCH "/api/me/bookshelf/$SMOKE_BOOK_ID" "{\"groupId\":$custom_group_id,\"pinned\":true}" "$TOKEN")"
  assert_status "move and pin bookshelf book" "$code" "200"

  code="$(request GET "/api/me/bookshelf?groupId=$custom_group_id&page=1&pageSize=10" "" "$TOKEN")"
  assert_status "custom bookshelf list" "$code" "200"

  code="$(request GET "/api/me/bookshelf/groups/$custom_group_id" "" "$TOKEN")"
  assert_status "bookshelf group detail" "$code" "200"

  code="$(request POST /api/me/bookshelf/batch "{\"action\":\"move\",\"groupId\":0,\"bookIds\":[$SMOKE_BOOK_ID]}" "$TOKEN")"
  assert_status "batch move bookshelf book" "$code" "200"

  code="$(request GET "/api/me/bookshelf?page=1&pageSize=10" "" "$TOKEN")"
  assert_status "root bookshelf list after move back" "$code" "200"

  code="$(request POST /api/me/bookshelf/batch "{\"action\":\"remove\",\"bookIds\":[$SMOKE_BOOK_ID]}" "$TOKEN")"
  assert_status "batch remove bookshelf book" "$code" "200"
}

login_admin() {
  admin_body="{\"username\":\"$SUPER_ADMIN_USERNAME\",\"password\":\"$SUPER_ADMIN_PASSWORD\"}"
  code="$(request POST /api/admin/auth/login "$admin_body")"

  if [ "$code" != "200" ]; then
    echo "SKIP admin login: admin login returned $code"
    return 1
  fi

  SUPER_ADMIN_TOKEN="$(json_value token)"
  ADMIN_TOKEN="$SUPER_ADMIN_TOKEN"
  if [ -z "$ADMIN_TOKEN" ]; then
    echo "SKIP admin login: admin login response did not contain token"
    return 1
  fi

  return 0
}

promote_smoke_user_to_author() {
  if [ -z "$SMOKE_USER_ID" ]; then
    echo "FAIL promote smoke user: missing smoke user id" >&2
    exit 1
  fi

  if [ -z "$ADMIN_TOKEN" ]; then
    if ! login_admin; then
      echo "FAIL promote smoke user: admin login failed" >&2
      exit 1
    fi
  fi

  code="$(request PATCH "/api/admin/users/$SMOKE_USER_ID/author-role" "" "$ADMIN_TOKEN")"
  assert_status "promote smoke user to author" "$code" "200"
}

check_user_authoring() {
  sample="$TMP_DIR/user-sample.txt"
  cat > "$sample" <<'TXT'
第一章 用户上传
这是普通用户上传的测试章节。

第二章 继续
普通用户继续更新。
TXT

  code="$(curl -sS -X POST "$BASE_URL/api/me/books/upload" \
    -H "Authorization: Bearer $TOKEN" \
    -F "title=Smoke User Novel" \
    -F "categoryId=$CATEGORY_ID" \
    -F "description=Smoke user upload sample" \
    -F "file=@$sample;type=text/plain" \
    -o "$TMP_DIR/response.json" \
    -w "%{http_code}")" || code="000"

  assert_status "user txt upload" "$code" "200,201"
  book_id="$(json_number bookId)"
  if [ -z "$book_id" ]; then
    echo "FAIL user txt upload: response does not contain bookId" >&2
    sed -n '1,120p' "$TMP_DIR/response.json" >&2 || true
    exit 1
  fi
  SMOKE_BOOK_ID="$book_id"

  code="$(request GET "/api/me/books?page=1&pageSize=5" "" "$TOKEN")"
  assert_status "my book list" "$code" "200"

  update_body="{\"title\":\"Smoke User Managed Novel\",\"categoryId\":$CATEGORY_ID,\"description\":\"Updated by user smoke\"}"
  code="$(request PATCH "/api/books/$book_id" "$update_body" "$TOKEN")"
  assert_status "user update own book" "$code" "200"

  add_body="{\"title\":\"第三章 用户新增\",\"content\":\"这是普通用户新增章节。\"}"
  code="$(request POST "/api/books/$book_id/chapters" "$add_body" "$TOKEN")"
  assert_status "user add own chapter" "$code" "200"
  chapter_id="$(json_number id)"
  if [ -z "$chapter_id" ]; then
    echo "FAIL user add own chapter: response does not contain id" >&2
    sed -n '1,120p' "$TMP_DIR/response.json" >&2 || true
    exit 1
  fi

  check_avatar_upload
  check_bookshelf_management

  code="$(request DELETE "/api/books/$book_id/chapters/$chapter_id" "" "$TOKEN")"
  assert_status "user delete own chapter" "$code" "200"

  code="$(request DELETE "/api/books/$book_id" "" "$TOKEN")"
  assert_status "user delete own book" "$code" "200"
}

prepare_admin() {
  if ! login_admin; then
    echo "SKIP admin categories: admin login failed"
    return 1
  fi
  return 0
}

check_admin_categories() {
  name="Smoke分类$SMOKE_USERNAME"
  body="{\"name\":\"$name\"}"
  code="$(request POST /api/admin/categories "$body" "$ADMIN_TOKEN")"
  assert_status "admin create category" "$code" "200"
  category_id="$(json_number id)"
  if [ -n "$category_id" ]; then
    body="{\"name\":\"${name}改\"}"
    code="$(request PATCH "/api/admin/categories/$category_id" "$body" "$ADMIN_TOKEN")"
    assert_status "admin update category" "$code" "200"
    code="$(request DELETE "/api/admin/categories/$category_id" "" "$ADMIN_TOKEN")"
    assert_status "admin delete unused category" "$code" "200"
  fi
}

main() {
  need_curl
  mkdir -p "$TMP_DIR"

  echo "Smoke target: $BASE_URL"
  check_health
  check_public_books
  register_and_login
  check_account_management
  promote_smoke_user_to_author
  relogin_smoke_user
  check_user_authoring
  prepare_admin
  if [ -n "$ADMIN_TOKEN" ]; then
    check_admin_categories
  fi
}

main "$@"
