# `testcase-extractor`
Process Builder Initiative

A `testcase-extractor` extracts test cases for FEEL expressions from
[Decision Model and Notation (DMN) Technology
Compatibility Kit (TCK)](https://github.com/dmn-tck/tck) test suite.

## Description

`testcase-extractor` finds all feel related test cases (based on test case name)
and extract them to YAML file.

## Usage

```
Usage of ./testcase-extractor:
  -dir string
        Directory to start search for DMN model files and Test Cases files. (default ".")
  -output-file string
        File to save output to. Default is "-" (STDOUT). (default "-")
```

### How to Use

1. Get DMN TCK test cases from https://github.com/dmn-tck/tck and save them to
local dir, e.g. `./resources`
   ```
   git clone https://github.com/dmn-tck/tck ./resources
   ```
2. Run `testcase-extractor`
   ```
   testcase-extractor --dir ./resources
   ```
3. Enjoy testcases configuration in YAML format. Example output excerpt below:
   ```
   ...
   - model:
      dir: /TestCases/compliance-level-3/1149-feel-today-function
      name: 1149-feel-today-function.dmn
    test-cases:
      - id: "001"
        description: Will give date
        tests:
          - feel-expression: today() instance of date
            expected:
              result:
                  value: "true"
                  type: boolean
      - id: "002"
        description: Too many params gives null
        tests:
          - feel-expression: today(123)
            expected:
              result:
                  value: null
   
   2006 test cases for feel expressions extracted
   158 TCK DMN and Test Case files processed
   It took 99.604216ms
   ```
   
Or use `make extract-testcases` in repo root.
