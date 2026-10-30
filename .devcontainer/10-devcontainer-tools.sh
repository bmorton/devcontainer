# Installed by .devcontainer/Dockerfile: mise shims and Go/Cargo bin dirs.
export MISE_DATA_DIR=/usr/local/share/mise MISE_TRUSTED_CONFIG_PATHS=/etc/mise
export GOPATH=/go RUSTUP_HOME=/usr/local/rustup CARGO_HOME=/usr/local/cargo
for _d in "$CARGO_HOME/bin" "$GOPATH/bin" "$MISE_DATA_DIR/shims"; do
  case ":$PATH:" in *":$_d:"*) ;; *) PATH="$_d:$PATH" ;; esac
done
unset _d
export PATH
