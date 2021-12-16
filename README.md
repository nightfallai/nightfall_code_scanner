# nightfall_code_scanner

![Nightfall_Code_Scanner](https://cdn.nightfall.ai/nightfall-dark-logo-tm.png "Nightfall_Code_Scanner")

### Nightfall_Code_Scanner - a code review tool that protects you from committing sensitive information

The `nightfall_code_scanner` is a code review tool that protects you from committing sensitive information to your
version control. It is intended to be used as a part of your CI to simplify the development process, improve your
security, and ensure you never accidentally leak secrets or other sensitive information via an accidental commit.

## Supported Services
* [GitHub Action](https://github.com/nightfallai/nightfall_dlp_action)
* [CircleCI Orb](https://github.com/nightfallai/nightfall_circle_orb)

## NightfallDLP Config File

The `.nightfalldlp/config.json` file is a centralized config file that defines which detectors to use when scanning
content from pull requests. It supports the following configurations:

### Detection Rule UUIDs

A Detection Rule UUID is a unique identifier for a [Detection Rule](https://docs.nightfall.ai/docs/entities-and-terms-to-know#detection-rules)
that has been built in the [Nightfall Dashboard](https://app.nightfall.ai/). Users can copy a set of UUIDs from
the dashboard, then paste them into the config file, like below:

```json
{ "detectionRuleUUIDs": ["A0BA0D76-78B4-4E10-B653-32996060316B", "c035c4f3-eeb2-4764-a715-c8461f388661"] }
```

### Detection Rules

Detection Rules are a list of detectors specified inline. In addition to the list of inline detector definitions, a
detection rule contains an optional display name, as well as an operator `logicalOp` that is applied to the list
of detectors.

```json
{
  "detectionRules": [
    {
      "name": "my detection rule",
      "logicalOp": "ANY",
      "detectors": []
    }
  ]
}
```

The `logicalOp` field supports two values: `ANY` (logical OR), and `ALL` (logical AND).
Consider scanning the payload `"my ssn is 678-99-8212"` against a detection rule containing the `CREDIT_CARD_NUMBER` and
`US_SOCIAL_SECURITY_NUMBER` detectors with the logical op `ANY`. In this case, Nightfall would return a finding for the
string `678-99-8212` from the SSN detector. However, scanning this same payload with a logical op of `ALL` will not
return any findings, since the string does not contain any credit card numbers.

### Detector

A detector represents the atomic unit of scanning for sensitive data. The simplest examples of detectors leverage
Nightfall's pre-built detectors for data types such as `CREDIT_CARD_NUMBER` or `API_KEY`. The list of available
detectors is available [here](https://docs.nightfall.ai/docs/detector-glossary). As an example:

```json
  {
    "detector": {
      "detectorType": "NIGHTFALL_DETECTOR",
      "nightfallDetector": "API_KEY",
      "displayName": "apiKeyDetector",
      "minNumFindings": 1,
      "minConfidence": "LIKELY"
    }
  }
```

Two common configurations across all types of detectors are:
- `minNumFindings`: the number of findings that must match the input string in order for Nightfall to return a finding
- `minConfidence`: the minimum confidence threshold that Nightfall must meet in order to return a finding.

## Additional Configuration

In addition to detection rule configuration, you can also use your config file to further customize the scan process.

### Token Exclusion

To ignore specific tokens, you can add the `tokenExclusionList` field to your config. The `tokenExclusionList` is a
list of strings, and it mutes findings that match any of the provided regular expression patterns.

Here's an example use case:

```
tokenExclusionList: ["NF-gGpblN9cXW2ENcDBapUNaw3bPZMgcABs", "^127\\."]
```

In the example above, findings with the API token `NF-gGpblN9cXW2ENcDBapUNaw3bPZMgcABs`, as well as
IP addresses starting with `127.`, would not be reported. For more information on how we match tokens, take a
look at [the docs](https://docs.nightfall.ai/docs/entities-and-terms-to-know#custom-detectors).

### File Inclusion/Exclusion

The field `fileExclusionList` specifies glob patterns for files that should not be scanned during CI.
Conversely, the field `fileInclusionList` specifies glob patterns for the files that should be scanned during CI.

Here's an example use case:

```
    fileExclusionList: ["*/tests/*"],
    fileInclusionList: ["*.go", "*.json"]
```

In the example, we are ignoring all file paths with a `tests` subdirectory, and only scanning on `go` and `json` files.

## Configuration Examples

- Using a pre-built Detection Rule

```json
{ "detectionRules": ["83533b7c-de88-466a-b137-fceb8f2a8a57"] }
```

- Inline Detection Rule using Nightfall Detectors

```json
{
  "detectionRules": [
    {
      "name": "my rule",
      "detectors": [
        {
          "minNumFindings": 1,
          "minConfidence": "POSSIBLE",
          "displayName": "nfAPIKEY",
          "detectorType": "NIGHTFALL_DETECTOR",
          "nightfallDetector": "API_KEY"
        },
        {
          "minNumFindings": 2,
          "minConfidence": "POSSIBLE",
          "displayName": "nfCC",
          "detectorType": "NIGHTFALL_DETECTOR",
          "nightfallDetector": "CREDIT_CARD_NUMBER"
        }
      ],
      "logicalOp": "ANY"
    }
  ]
}
```

- Inline Detection Rules using Regex and WordList detectors

```json
{
  "detectionRules": [
    {
      "name": "regex rule",
      "detectors": [
        {
          "minNumFindings": 1,
          "minConfidence": "POSSIBLE",
          "displayName": "simpleRegexCouponDetector",
          "detectorType": "REGEX",
          "regex": {
            "pattern": "coupon:\\d{4,}",
            "isCaseSensitive": false
          }
        }
      ],
      "logicalOp": "ANY"
    },
    {
      "name": "word list rule",
      "detectors": [
        {
          "minNumFindings": 1,
          "minConfidence": "POSSIBLE",
          "displayName": "simpleWordListKeyDetector",
          "detectorType": "WORD_LIST",
          "wordList": {
            "values": ["key", "credential"],
            "isCaseSensitive": false
          }
        }
      ],
      "logicalOp": "ANY"
    }
  ]
}
```

- Detection Rule containing context and exclusion rules:

```json
{
  "detectionRules": [
    {
      "name": "",
      "detectors": [
        {
          "minNumFindings": 2,
          "minConfidence": "POSSIBLE",
          "displayName": "nfCC",
          "detectorType": "NIGHTFALL_DETECTOR",
          "nightfallDetector": "CREDIT_CARD_NUMBER",
          "contextRules": [
            {
              "regex": {
                "pattern": "credit card",
                "isCaseSensitive": true
              },
              "proximity": {
                "windowBefore": 30,
                "windowAfter": 30
              },
              "confidenceAdjustment": {
                "fixedConfidence": "VERY_LIKELY"
              }
            }
          ],
          "exclusionRules": [
            {
              "matchType": "PARTIAL",
              "exclusionType": "REGEX",
              "regex": {
                "pattern": "4242-4242-4242-4242",
                "isCaseSensitive": true
              },
              "wordList": null
            }
          ]
        }
      ],
      "logicalOp": "ANY"
    }
  ],
  "maxNumberConcurrentRoutines": 5,
  "tokenExclusionList": [
    "4916-6734-7572-5015",
    "301-123-4567",
    "1-240-925-5721",
    "xG0Ct4Wsu3OTcJnE1dFLAQfRgL6b8tIv"
  ],
  "fileInclusionList": ["*"]
}
```
