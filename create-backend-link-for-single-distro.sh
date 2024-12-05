cd backend/
for item in *; do
  if [[ "$item" != "main.go" && "$item" != "config.toml" ]]; then
    ln -s "$PWD/$item" "../single-distro/$item"
  fi
done