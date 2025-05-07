printf 'Waiting for GitLab container to become healthy\n'

while true ; do
  if [ "$(curl http://127.0.0.1:9999/users/sign_in -o /dev/null --silent -w "%{http_code}")" -eq 200 ] ; then
    break
  fi
  echo "Gitlab is not yet healthy..."
  sleep 5
done

echo
echo "GitLab is healthy...!"
