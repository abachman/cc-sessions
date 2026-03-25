## Branch naming

Git branch names must include a `descriptive-name`: three to five words of descriptive text summarizing what the branch is for, lowercase, with dash separators.

## Pull requests

When creating pull requests on GitHub be brief and concise, use high-level technical language, summarize the impact of the changes, and give a short explanation of how to test the changes.

## Branch protections

`main` is a protected branch. Do not attempt to merge PRs, make commits, or push changes to `main`. If you are working on a branch that already has a PR, it will be reviewed by humans and merged by hand, it is fine to just use `git push` to make sure remote is up to date.

## gh command line tool

You can use the `gh` command line tool to create PRs, check on the progress of CI workflows, and check out pull requests for review.
