mkdir -p single-distro/frontend
for file in frontend/dist; do
  ln -s "$(realpath "$file")" "single-distro/frontend/$(basename "$file")"
done
