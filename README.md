# nightfall_code_scanner

![Nightfall_Code_Scanner](https://cdn.nightfall.ai/nightfall-dark-logo-tm.png "Nightfall_Code_Scanner")

### Nightfall_Code_Scanner - a code review tool that protects you from committing sensitive information

nightfall_code_scanner scans your code for secrets or sensitive information. It’s intended to be used as a part of your CI to simplify the development process, improve your
security, and ensure you never accidentally leak secrets or other sensitive information via an accidental commit.

## Supported Services

### GithubActions

[nightfall_dlp_action](https://github.com/nightfallai/nightfall_dlp_action)

## Nightfalldlp Config File

The .nightfalldlp/config.json file is used as a centralized config file to control what conditions/detectors and content you want to scan for pull requests. It includes following fields available to adjust to fit your needs.

### ConditionSetUUID

A condition set uuid is a unique identifier of a condition set, also called policy, which can be defined through app.nightfall.ai UI.
Once defined, you can only input the uuid in the your config file, e.g.

```json
{ "conditionSetUUID": "A0BA0D76-78B4-4E10-B653-32996060316B" }
```

Note: by default, if both conditionSetUUID and conditions are specified, only conditionSetUUID will be used.

### Conditions

Conditions are a list of conditions specified inline. Each condition has a detector struct within and two extra parameters of minNumFindings and minConfidence as below.

```json
{
  "conditions": [
    {
      "minNumFindings": 1,
      "minConfidence": "Possible",
      "detector": {}
    }
  ]
}
```

minNumFindings specify minimal number of findings to trigger findings to return for one request, e.g. if you set minNumFindings to be 2, and only 1 finding within the request payload, then this finding will be filtered.

minConfidence specifies the minimal threshold to trigger a potential finding to be captured, in total we have five levels of confidence from least sensitive to most sensitive:

- VERY_LIKELY
- LIKELY
- POSSIBLE
- UNLIKELY
- VERY_UNLIKELY

### Detector

A detector is either a prebuilt detector from nightfall or customized regex | wordlist by customer, specified by
detectorType field.

- nightfall prebuilt detector

  ```json
  {
    "detector": {
      "detectorType": "NIGHTFALL_DETECTOR",
      "nightfallDetector": "API_KEY",
      "displayName": "credit card detector"
    }
  }
  ```

  Within detector struct

  - First specify detectorType as NIGHTFALL_DETECTOR
  - Pick the nightfall detector you want from the list
    - API_KEY
    - CRYPTOGRAPHIC_KEY
    - RANDOMLY_GENERATED_TOKEN
    - CREDIT_CARD_NUMBER
    - US_SOCIAL_SECURITY_NUMBER
    - AMERICAN_BANKERS_CUSIP_ID
    - US_BANK_ROUTING_MICR
    - ICD9_CODE
    - ICD10_CODE
    - US_DRIVERS_LICENSE_NUMBER
    - US_PASSPORT
    - PHONE_NUMBER
    - IP_ADDRESS
    - EMAIL_ADDRESS
  - Put a display name for your detector, as this will be attached on your findings

- customized regex

  For convenience, we also support customized regex as a detector, which you can fill in as follow.

  ```json
  {
    "detector": {
      "detectorType": "REGEX",
      "regex": {
        "pattern": "coupon",
        "isCaseSensitive": false
      },
      "displayName": "simpleRegexCouponDetector"
    }
  }
  ```

- word list

  In case you are not famaliar what regex is, we also support word list searching, simply put in a list of words you want to find.

  ```json
  {
    "detector": {
      "detectorType": "WORD_LIST",
      "wordList": {
        "values": ["key", "credential"],
        "isCaseSensitive": false
      },
      "displayName": "simpleWordListKeyDetector"
    }
  }
  ```

- [Extra Parameters Within Detector]

  Besides to define which detector to call, we also allow you to optionally specify some rules you want to attach surround the findinds and finding themselves. They are contextRules and exclusionRules.

  - contextRules
    A context rule is defined as the surrounding context(pre/post chars) of a finding, you can define a trigger/rule in this section. Once it's trigger, the confidence of finding will be adjusted accordingly base on your confid.

    Example:

    ```json
    {
      "detector": {
        // ...... other detector fileds
        "contextRules": [
          {
            "regex": {
              "pattern": "test cc",
              "isCaseSensitive": true
            },
            "proximity": {
              "windowBefore": 30,
              "windowAfter": 30
            },
            "confidenceAdjustment": {
              "fixedConfidence": "VERY_UNLIKELY"
            }
          }
        ]
      }
    }
    ```

    - regex field specifies a regex to trigger
    - proximity specifies how many pre|post chars surround findings we want to do the search
    - confidenceAdjustment specifies what's the confidence you want to change for the finding once trigger the rule

    In this example, if we have real test like test cc: 4242-4242-4242-4242, and 4242-4242-4242-4242 is detected as a credit card number with confidence of POSSIBLE. After we applied such context rules, since the pre chars test cc matches the regex, the confidence of such findings will drop down to VERY_UNLIKELY as specified

  - exclusionRules
    Similar to context rules, you can also apply rules on findings themselves, in case you find certain findings or patterns appear to be super noisy in your case. To deactivate such appearance, you can do

    Example:

    ```json
    {
      "detector": {
        // ...... other detector fileds
        "exclusionRules": [
          {
            "matchType": "PARTIAL",
            "exclusionType": "REGEX",
            // specify one of these values base on type specified above
            "regex": {
              "pattern": "4242-4242-4242-4242",
              "isCaseSensitive": true
            },
            "wordList": {
              "values": ["4242", "1234"],
              "isCaseSensitive": false
            }
          }
        ]
      }
    }
    ```

    - exclusionType specifies either a REGEX or WORD_LIST, similar to how you define a customized regex or word list detector
    - regex field specifies a regex to trigger, if you choose to use REGEX type
    - matchType could be either PARTIAL or FULL, PARTIAL means the regex only matches part of the finding, maybe first 5 chars, and with PARTIAL specified, we'd deactivate such findings as well, FULL means the regex has to match the whole finding context to deactivate
      Suppose we have a finding of "4242-4242" with exclusion regex of 4242, if you use PARTIAL, this finding will be deactivated, while FULL then not, since the regex only matches partial of the finding

## Additional Configuration

Aside from which conditions to scan on, you can add additional fields to your config, `./nightfall/config.json`, to ignore tokens and files as well as specify which files to exclusively scan on.

### Token Exclusion

To ignore specific tokens, you can add the `tokenExclusionList` field to your config. The `tokenExclusionList` is a list of strings, and it mutes findings that match any of the given regex patterns.

Here's an example use case:

`tokenExclusionList: ["NF-gGpblN9cXW2ENcDBapUNaw3bPZMgcABs", "^127\\."]`

In the example above, findings with the API token `NF-gGpblN9cXW2ENcDBapUNaw3bPZMgcABs` as well as local IP addresses starting with `127.` would not be reported. For more information on how we match tokens, take a look at [regexp](https://golang.org/pkg/regexp/).

### File Inclusion/Exclusion

To omit files from being scanned, you can add the `fileExclusionList` field to your config. In addition, to only scan on certain files, add the `fileInclusionList` to the config.

Here's an example use case:

```

    fileExclusionList: ["*/tests/*"],
    fileInclusionList: ["*.go", "*.json"]

```

In the example, we are ignoring all file paths with a `tests` subdirectory, and only scanning on `go` and `json` files.
Note: we are using [gobwas/glob](https://github.com/gobwas/glob) to match file path patterns. Unlike the token regex matching, file paths must be completely matched by the given pattern. e.g. If `tests` is a subdirectory, it will not be matched by `tests/*`, which is only a partial match.

## Extra Real Examples

To summarize, we provide sevaral more examples as below

- Config conditionSet through app.nightfall.ai UI

```json
{ "conditionSetUUID": "UUID HERE" }
```

- Config conditions inline with prebuilt detector

```json
{
  "conditions": [
    {
      "minNumFindings": 1,
      "minConfidence": "POSSIBLE",
      "detector": {
        "detectorType": "NIGHTFALL_DETECTOR",
        "nightfallDetector": "API_KEY",
        "displayName": "nfAPIKEY"
      }
    },

    {
      "minNumFindings": 2,
      "minConfidence": "POSSIBLE",
      "detector": {
        "detectorType": "NIGHTFALL_DETECTOR",
        "nightfallDetector": "CREDIT_CARD_NUMBER",
        "displayName": "nfCC"
      }
    }
  ]
}
```

- Config conditions inline with your own regex | word list

```json
{
  "conditions": [
    {
      "minNumFindings": 1,
      "minConfidence": "POSSIBLE",
      "detector": {
        "detectorType": "REGEX",
        "regex": {
          "pattern": "coupon",
          "isCaseSensitive": false
        },
        "displayName": "simpleRegexCouponDetector"
      }
    },
    {
      "minNumFindings": 1,
      "minConfidence": "POSSIBLE",
      "detector": {
        "detectorType": "WORD_LIST",
        "wordList": {
          "values": ["key", "credential"],
          "isCaseSensitive": false
        },
        "displayName": "simpleWordListKeyDetector"
      }
    }
  ]
}
```

- Detailed configuration, if you find there're noisy results you want to get rid of.

```json
{
  "conditions": [
    {
      "minNumFindings": 2,
      "minConfidence": "POSSIBLE",
      "detector": {
        "detectorType": "NIGHTFALL_DETECTOR",
        "nightfallDetector": "CREDIT_CARD_NUMBER",
        "displayName": "nfCC"
      },
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
          }
        }
      ]
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
