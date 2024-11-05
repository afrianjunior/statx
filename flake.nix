{
  description = "A Nix Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:

  utils.lib.eachDefaultSystem (system:
    let goVersion = 22;
        dbLocation = "sqlite3://data/sqlite/statx.db";
        overlays = [ (final: prev: { go = prev."go_1_${toString goVersion}"; }) ];
        pkgs = import nixpkgs { inherit overlays system; };
    
    scripts = with pkgs; [
      (writeScriptBin "create-migration" ''
        migrate create -ext sql -dir migrations -seq $@
      '')

      (writeScriptBin "run-migration" ''
        migrate -path migrations -database ${dbLocation} up
      '')

    (writeScriptBin "drop-migration" ''
      migrate -path migrations -database ${dbLocation} drop
    '')


      (writeScriptBin "rollback" ''
        migrate -path migrations -database ${dbLocation} down
      '')
    ];
    in
    rec {
      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go

          go-migrate

          gopls

          gotools

          golangci-lint
        ] ++ scripts;
      };

      shellHook = ''
        # allow for external dependency to be
        # install on nix os imperatively 
        # using `go install` when certain 
        # dependencies not available/outdated 
        # on nixpkgs
        export GOPATH="$(${pkgs.go}/bin/go env GOPATH)"
        export PATH="$PATH:$GOPATH/bin"

        ${pkgs.go}/bin/go version
      '';
    });

}
