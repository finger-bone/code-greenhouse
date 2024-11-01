mkdir -p single-distro
for file in backend/*; do
  if [[ $(basename "$file") != "main.go" && $(basename "$file") != "config.toml" ]]; then
    target="single-distro/$(basename "$file")"
    if [[ ! -e "$target" ]]; then
      ln -s "$(realpath "$file")" "$target"
    fi
  fi
done