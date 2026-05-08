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
ADMIN_USERNAME="${ADMIN_USERNAME:-$(env_value ADMIN_USERNAME)}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-$(env_value ADMIN_PASSWORD)}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin_password_change_me}"
TMP_DIR="${TMPDIR:-/tmp}/novel-reader-smoke.$$"

TOKEN=""
ADMIN_TOKEN=""
CATEGORY_ID=""

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

  code="$(request DELETE "/api/books/$book_id/chapters/$chapter_id" "" "$TOKEN")"
  assert_status "user delete own chapter" "$code" "200"

  code="$(request DELETE "/api/books/$book_id" "" "$TOKEN")"
  assert_status "user delete own book" "$code" "200"
}

try_admin_upload() {
  admin_body="{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}"
  code="$(request POST /api/auth/login "$admin_body")"

  if [ "$code" != "200" ]; then
    echo "SKIP admin upload: admin login returned $code"
    return
  fi

  ADMIN_TOKEN="$(json_value token)"
  if [ -z "$ADMIN_TOKEN" ]; then
    echo "SKIP admin upload: admin login response did not contain token"
    return
  fi

  sample="$TMP_DIR/sample.txt"
  cat > "$sample" <<'TXT'
第一章 开始
这是 smoke check 上传的测试章节。

第二章 继续
这是第二章内容。
TXT

  code="$(curl -sS -X POST "$BASE_URL/api/admin/books/upload" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -F "title=Smoke Test Novel" \
    -F "categoryId=$CATEGORY_ID" \
    -F "description=Smoke check upload sample" \
    -F "file=@$sample;type=text/plain" \
    -o "$TMP_DIR/response.json" \
    -w "%{http_code}")" || code="000"

  assert_status "admin txt upload" "$code" "200,201"
  book_id="$(json_number bookId)"
  if [ -n "$book_id" ]; then
    check_admin_management "$book_id"
  fi
}

check_admin_management() {
  book_id="$1"

  code="$(request GET "/api/admin/books?page=1&pageSize=5" "" "$ADMIN_TOKEN")"
  assert_status "admin book list" "$code" "200"

  update_body="{\"title\":\"Smoke Managed Novel\",\"categoryId\":$CATEGORY_ID,\"description\":\"Updated by smoke\"}"
  code="$(request PATCH "/api/admin/books/$book_id" "$update_body" "$ADMIN_TOKEN")"
  assert_status "admin update book" "$code" "200"

  add_body="{\"title\":\"第三章 新增\",\"content\":\"这是 smoke 新增章节。\"}"
  code="$(request POST "/api/admin/books/$book_id/chapters" "$add_body" "$ADMIN_TOKEN")"
  assert_status "admin add chapter" "$code" "200"
  chapter_id="$(json_number id)"
  if [ -z "$chapter_id" ]; then
    echo "FAIL admin add chapter: response does not contain id" >&2
    sed -n '1,120p' "$TMP_DIR/response.json" >&2 || true
    exit 1
  fi

  patch_body="{\"title\":\"第三章 已修改\",\"content\":\"这是 smoke 修改后的章节。\"}"
  code="$(request PATCH "/api/admin/books/$book_id/chapters/$chapter_id" "$patch_body" "$ADMIN_TOKEN")"
  assert_status "admin update chapter" "$code" "200"

  code="$(request DELETE "/api/admin/books/$book_id/chapters/$chapter_id" "" "$ADMIN_TOKEN")"
  assert_status "admin delete chapter" "$code" "200"

  code="$(request DELETE "/api/admin/books/$book_id" "" "$ADMIN_TOKEN")"
  assert_status "admin delete book" "$code" "200"

  code="$(request GET "/api/books/$book_id")"
  assert_status "deleted book not public" "$code" "404"
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
  check_user_authoring
  try_admin_upload
  if [ -n "$ADMIN_TOKEN" ]; then
    check_admin_categories
  fi
}

main "$@"
