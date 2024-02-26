# MLS DISPLAY RULES

## Executing Integration Tests
- Run ```run-test.sh``` file to execute all tests
- To update the test search for necessary ```*_test.go``` file and edit as required
- The tests are run through docker compose, configured in ```docker-compose.test.yml```
- Test Data is stored in ```test_data.json```. If you add new data, make sure to update the test cases that validated them, as necessary

## Wiki
- Wiki URL TODO

## Functionality
- API for Mls Display Rules.

## Executing in local environment 
- Execute ```go run main.go```