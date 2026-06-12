#!/usr/bin/env python3
"""Headlessly create a persistent default Secret Service collection.

gnome-keyring's spec-compliant CreateCollection requires gcr-prompter
(GTK/Wayland), which devcontainers don't have. This script calls the
non-spec InternalUnsupportedGuiltRiddenInterface.CreateWithMasterPassword
method, which accepts the master password inline. Read from
KEYRING_PASSWORD env (empty by default).

Idempotent: exits 0 if the 'default' alias already resolves.
"""
from __future__ import annotations
import os, sys
from jeepney import DBusAddress, new_method_call
from jeepney.io.blocking import open_dbus_connection

SP, SB = "/org/freedesktop/secrets", "org.freedesktop.secrets"
SS = "org.freedesktop.Secret.Service"
GI = "org.gnome.keyring.InternalUnsupportedGuiltRiddenInterface"


def main() -> int:
    conn = open_dbus_connection(bus="SESSION")
    svc = DBusAddress(SP, bus_name=SB, interface=SS)

    reply = conn.send_and_get_reply(new_method_call(svc, "ReadAlias", "s", ("default",)))
    if reply.body and reply.body[0] != "/":
        print(f"[keyring-bootstrap] default alias already set to {reply.body[0]}")
        return 0

    # 'plain' session: master password travels in cleartext over the
    # local unix socket. That's fine because the password is empty (or
    # an environment variable already inside the container).
    reply = conn.send_and_get_reply(
        new_method_call(svc, "OpenSession", "sv", ("plain", ("s", "")))
    )
    _, session_path = reply.body

    password = os.environ.get("KEYRING_PASSWORD", "").encode("utf-8")
    master = (session_path, b"", password, "text/plain")
    attrs  = {"org.freedesktop.Secret.Collection.Label": ("s", "Login")}

    guilt = DBusAddress(SP, bus_name=SB, interface=GI)
    reply = conn.send_and_get_reply(
        new_method_call(guilt, "CreateWithMasterPassword",
                        "a{sv}(oayays)", (attrs, master))
    )
    if reply.header.message_type.name != "method_return":
        print(f"[keyring-bootstrap] CreateWithMasterPassword failed: {reply.body}", file=sys.stderr)
        return 1
    coll_path = reply.body[0]
    print(f"[keyring-bootstrap] created collection {coll_path}")

    reply = conn.send_and_get_reply(
        new_method_call(svc, "SetAlias", "so", ("default", coll_path))
    )
    if reply.header.message_type.name != "method_return":
        print(f"[keyring-bootstrap] SetAlias failed: {reply.body}", file=sys.stderr)
        return 1
    print("[keyring-bootstrap] bound 'default' alias")
    return 0


if __name__ == "__main__":
    sys.exit(main())
