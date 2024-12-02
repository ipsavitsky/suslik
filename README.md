# Gitlab Gopher

This bot exists to perform review roulette in GitLab.

## Isn't it already part of GitLab?

Surprisingly, no

In [this](https://about.gitlab.com/blog/2019/10/23/reviewer-roulette-one-year-on/) blogpost from 2019 GitLab writes:

> The next steps are to turn Reviewer Roulette into a feature that all users of GitLab can benefit from, perhaps by leveraging the CODEOWNERS file.

Searching the internet or GitLab handbook on existence of such feature yields no results. So, for now, this bot exists as a workaround.

## Configuration

Right now, Gitlab Gopher supports two configuration options:

``` toml
# Set your GitLab instance token
token = "<token>"
# Set the base URL for the GitLab instance
baseURL = "gitlab.example.com/api/v4"
# Set the branch that the bot will pick per-repo config from
reviewerFileRef = "main"
```

After that, the bot expects a `reviewers.yaml` file that will look like this:

``` yaml
# Set how many reviewers are expected in a typical review
reviewThreshold: 2
usernames:
  - reviewer_1
  - reviewer_2
```

> Why is this not a CODEOWNERS file?

This is the goal, but I was too lazy to implement complex logic of a `CODEOWNERS` file.

After deploying and putting all configuration in place, assign the bot as a reviewer to merge request. The bot will unassign itself, select random reviewers out of the list and assign them. If some reviewers were already assigned, they will not be unassigned, but instead the bot will assign reviewers up to the threshold. For example, if 1 reviewer is alread assigned and the threshold is set to 3, 2 additional reviewers will be randomly assigned.
