## arch-log

`arch-log` is a small program that displays the commit messages of Arch packages. It queries both Arch's central repo and the AUR.

It is available on AUR: https://aur.archlinux.org/packages/arch-log

### Rationale
If you have multiple custom packages in Arch, you know the drag: You notice that some package of which you have a custom fork (or just an AUR package with long compile time) has a new version -- but only the pkgrel has changed.

The question then is: Do I need to rebuild / rebase / ... or can I ignore the change. To make this decision, it is necessary to have the package's changelog in quick access.

As I'm tired of clicking through different web interfaces, and I don't know of any other tool that provides this: `arch-log` was born.

### What does it do?

 1. Query https://archlinux.org/packages for the `pkgbase`.
2. If found: Query https://gitlab.archlinux.org (using Gitlab's REST API) for the commit and tag data.
3. Query https://aur.archlinux.org/rpc for `pkgbase`.
4. If found: Query https://aur.archlinux.org/cgit/aur.git (using the Atom Feed) for the commit data.

### What's with the name?

`paclog` was already taken.

### How does it look like?

![Example](https://necoro.dev/data/example_arch-log.png)
