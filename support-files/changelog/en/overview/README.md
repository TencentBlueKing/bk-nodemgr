# Version overviews

This directory stores independently authored English overviews for base versions such as `v3.0.1`. All `v3.0.1-alpha.*`, `v3.0.1-beta.*`, and `v3.0.1-release.*` iterations share that overview.

- Use one `<version>.md` file per base version, for example `v3.0.1.md`. Update this file in place as the overview evolves; neither the filename nor the overview listing carries a publication date.
- Describe the version's overall capabilities, highlights, and relevant usage or compatibility boundaries rather than copying individual iteration changes. Use standard Markdown headings, lists, and links; executable HTML and scripts are not supported content.
- Keep the same filename and supported facts in `../../zh/overview/` for the Chinese edition.
- Keep detailed iteration changelogs in the parent directory. An overview and a stable-release changelog with the same version are separate documents.
- Files are embedded by `support-files/changelog/embed.go`; rebuild and deploy Application after editing. No database or Wiki setting is involved.
- The UI opens **Version overview** for the running base version. If its document is absent in the selected language, an empty state is shown instead of substituting another version or aggregating iteration notes. Other available overviews remain selectable; **Update details** retains the existing iteration history.

This README is a maintenance guide and is not listed as a version overview.
