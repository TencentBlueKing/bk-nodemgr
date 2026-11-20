import argparse
import copy
import sys

import yaml


def load_definition(path):
    """Read a gateway definition without rendering deployment credentials."""
    with open(path, encoding="utf-8") as source:
        definition = yaml.safe_load(source)
    if not isinstance(definition, dict) or definition.get("spec_version") != 2:
        raise ValueError("MCP synchronization requires spec_version: 2")
    if not isinstance(definition.get("stages"), list):
        raise ValueError("definition stages must be a list")
    return definition


def prepare_definition(definition, product_path, enable_builtin):
    """Resolve complete stage overrides before falling back to product MCPs."""
    builtin = None
    stages = []
    for stage in definition["stages"]:
        if "mcp_servers" in stage:
            servers = stage["mcp_servers"]
        elif enable_builtin:
            if builtin is None:
                with open(product_path, encoding="utf-8") as source:
                    product = yaml.safe_load(source)
                if not isinstance(product, dict):
                    raise ValueError("product MCP catalog must be a mapping")
                builtin = product["mcp_servers"]
                if not isinstance(builtin, list) or not builtin:
                    raise ValueError("product mcp_servers must be a non-empty list")
            servers = copy.deepcopy(builtin)
            for server in servers:
                server["status"] = 1
        else:
            continue

        if not isinstance(servers, list):
            raise ValueError("stage mcp_servers must be a list; use [] to skip")
        if servers:
            stages.append({"name": stage["name"], "mcp_servers": servers})
    return {"spec_version": 2, "stages": stages}


def main():
    parser = argparse.ArgumentParser(description="Prepare stage MCP definitions")
    parser.add_argument("--definition", required=True)
    parser.add_argument("--mcp-servers", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--enable-builtin", action="store_true")
    args = parser.parse_args()

    try:
        definition = prepare_definition(
            load_definition(args.definition),
            args.mcp_servers,
            args.enable_builtin,
        )
        with open(args.output, "w", encoding="utf-8") as output:
            yaml.safe_dump(definition, output, allow_unicode=True)
    except (
        OSError,
        ValueError,
        KeyError,
        TypeError,
        yaml.YAMLError,
    ) as error:
        sys.exit("Failed to prepare MCP definition: {}".format(error))


if __name__ == "__main__":
    main()
