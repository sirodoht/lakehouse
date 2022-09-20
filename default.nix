with (import <nixpkgs> {});
mkShell {
  buildInputs = [
    postgresql_13
    golangci-lint
  ];
}
