GITLAB_ROOT_OAUTH=$(curl \
    --silent \
    --data-urlencode "grant_type=password" \
    --data-urlencode "username=root" \
    --data-urlencode "password=insecure1111" \
    localhost:9999/oauth/token | jq -r ".access_token")

GITLAB_ROOT_PAT=$(curl \
    --silent \
    --request POST \
    --header "Authorization: Bearer $GITLAB_ROOT_OAUTH" \
    --data "name=lorry" \
    --data "scopes[]=api,read_user,write_repository" \
    localhost:9999/api/v4/users/1/personal_access_tokens | jq -r ".token")

printf "setting suslik pat as %s\n" "$GITLAB_ROOT_PAT"

export SUSLIK_GITLAB_TOKEN="$GITLAB_ROOT_PAT"
