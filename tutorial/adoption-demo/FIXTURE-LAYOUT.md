# Adoption demo fixture layout

## Tracked source template

```text
tutorial/adoption-demo/fixture/
```

This directory contains the canonical synthetic Astra Relay starting state. It is versioned so the demo can be recreated deterministically.

Do not:

- open this directory as the Obsidian vault;
- run the tutorial inside it;
- treat it as Demon Docs or Space Rocks documentation;
- expect edits made elsewhere to update it automatically.

## Disposable working repository

Run one reset script from the Demon Docs checkout:

```bash
bash tutorial/adoption-demo/reset-demo.sh
```

```powershell
.\tutorial\adoption-demo\reset-demo.ps1
```

Both scripts create this sibling directory by default:

```text
../demon-docs-adoption-demo/
```

That sibling directory is the only intended tutorial workspace and Obsidian vault. It has its own repository boundary for Demon Docs initialization. Re-running the reset script deletes and recreates the sibling directory from the tracked source template.

The scripts refuse to create a target anywhere inside the Demon Docs checkout.
