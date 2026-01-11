# Codecov integration

This repository's CI uploads a coverage report artifact at `.ci/coverage.out`. If you want GitHub PR coverage checks and Codecov reports, follow these steps.

1. Create a Codecov account and add the repository (https://codecov.io/)
   - For public repositories Codecov can work without a token in many cases, but for private repos you need a repository upload token.

2. Create a repository token (if required)
   - In Codecov, go to the repository settings → *Repository Upload Token* and copy the token.

3. Add the Codecov token as a GitHub secret
   - Go to your repository on GitHub → Settings → Secrets and variables → Actions → New repository secret
   - Name: `CODECOV_TOKEN`
   - Value: the token you copied from Codecov

4. Verify the CI workflow
   - This repo's workflow already contains a Codecov step which will run only if `CODECOV_TOKEN` is present:

```yaml
      - name: Upload coverage to Codecov
        if: ${{ secrets.CODECOV_TOKEN != '' }}
        uses: codecov/codecov-action@v4
        with:
          files: .ci/coverage.out
          token: ${{ secrets.CODECOV_TOKEN }}
```

5. Configure Codecov behavior (optional)
   - You can add `codecov.yml` at the repository root to configure PR comments, status checks and thresholds. See the example `codecov.yml` in the repository.

6. Enable required checks (optional)
   - If you want PR merges to require coverage checks, add Codecov as a required status check in Branch Protection (Settings → Branches → main rule → Require status checks to pass). The Codecov check name will appear in the list after a successful upload.

Notes & tips
- For public repos Codecov often accepts uploads without a token; however, adding a token is recommended for tighter control and to enable protected settings.
- If you prefer not to store a token, you can still upload coverage artifacts with the workflow and inspect them manually.
- If you want PR comments disabled, set `comment: off` in `codecov.yml` or adjust the Codecov project settings.
