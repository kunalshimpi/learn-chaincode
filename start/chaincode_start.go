/*
Copyright IBM Corp 2016 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
		 http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	//"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/crypto/primitives"
)

// SimpleHealthChaincode example simple Chaincode implementation
type SimpleHealthChaincode struct {
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	primitives.SetSecurityLevel("SHA", 256)	
	err := shim.Start(new(SimpleHealthChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleHealthChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("**********Inside Init*******");
	if len(args) != 0 {
		return nil, errors.New("Incorrect number of arguments. Expecting 0")
	}
	err:=stub.CreateTable("InsuranceAmount", []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name:"Owner",Type: shim.ColumnDefinition_BYTES, Key: true},
		&shim.ColumnDefinition{Name:"Amount",Type:shim.ColumnDefinition_INT64, Key: false},
	})
	if err!= nil {
		return nil, errors.New("Error in Creating InsuranceAmount Table!")
	}

	adminCert, err := stub.GetCallerMetadata()

	if err!= nil{
		return nil, errors.New("Error Getting Metadata")
	}
	if len(adminCert) == 0 {
		return nil, errors.New("Admin Certificate is Empty!")
	}
	stub.PutState("admin", adminCert)

	fmt.Println("Admin is [%x] : ", adminCert)
	
	fmt.Println("Assigning Amount for admin!")
	_, err = stub.InsertRow("InsuranceAmount", shim.Row{
		Columns: []*shim.Column {
			&shim.Column{Value: &shim.Column_Bytes{Bytes:[]byte("admin")}},
			&shim.Column{Value: &shim.Column_Int64{Int64:1000}}},
	})
	if err != nil {
		return nil, errors.New("Failed to Assign Amount!")
	}
	

	fmt.Println("Init Finished!")

	return nil, nil
}
func (t *SimpleHealthChaincode) approve(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("approve is running ")
	
	if len(args) != 2{
		return nil, errors.New("Expected 2 arguments!")
	}

	ReqAmount, _ := strconv.ParseInt(args[1], 10, 64)

	applicant := []byte(args[0])
	/*if err != nil{
		return nil, errors.New("Decoding Failed!")
	}*/

	//adminCert, err := stub.GetState("admin")
	if err != nil{
		return nil, errors.New("Failed to get admin Certificate!")
	}

	/*ok, err := t.isCaller(stub, adminCert)
	if err != nil {
		return nil, errors.New("Failed to Check Certificates!")
	}
	if !ok {
		return nil, errors.New("Only Admin can call Approve function")
	}
*/
	fmt.Println("Assigning Amount!")
	
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_Bytes{Bytes:[]byte("admin")}}
	columns = append(columns, col1)

	row, err1 := stub.GetRow("InsuranceAmount", columns)
	if err1 != nil {
		return nil, errors.New("Cannot retrieve admin Row!")
	}
	BalAmount := row.Columns[1].GetInt64()
	BalAmount = BalAmount - ReqAmount
	
	_, err = stub.ReplaceRow("InsuranceAmount", shim.Row{
		Columns: []*shim.Column {
			&shim.Column{Value: &shim.Column_Bytes{Bytes:[]byte("admin")}},
			&shim.Column{Value: &shim.Column_Int64{Int64:BalAmount}},
		},
	})
	if err != nil {
		return nil, errors.New("Failed to Replace Admin Row!")
	}

	ok, err = stub.InsertRow("InsuranceAmount", shim.Row{
		Columns: []*shim.Column {
			&shim.Column{Value: &shim.Column_Bytes{Bytes:applicant}},
			&shim.Column{Value: &shim.Column_Int64{Int64:ReqAmount}},
			},
	})
	if err != nil {
		return nil, errors.New("Failed to Assign Amount!")
	}
	//???
	if !ok && err == nil {
		return nil, errors.New("Amount already Assigned")
	}

	fmt.Println("Approve Finished")
	return nil, err
}

func (t *SimpleHealthChaincode) transfer(stub shim.ChaincodeStubInterface, args []string)([]byte, error){
	if len(args) != 3{
		return nil, errors.New("Expected 3 arguments!")
	}	
	sender := []byte(args[0])
	receiver := []byte(args[1])
	transferAmount,_ := strconv.ParseInt(args[2], 10, 64)

//getowner asst from bc
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_Bytes{Bytes:sender}}
	columns = append(columns, col1)

	row, err := stub.GetRow("InsuranceAmount", columns)
	if err != nil {
		return nil, errors.New("Failed to Get sender's Balance amount")
	}
//chwck caller and Owner
	/*ok, err := t.isCaller(stub, sender)
	if err != nil {
		return nil, errors.New("Failed checking sender & caller identity")
	}
	if !ok {
		return nil, errors.New("The caller is not the owner of the amount")
	}*/
//change assets  of sender
	BalanceAmount := row.Columns[1].GetInt64()
	BalanceAmount = BalanceAmount - transferAmount
//replace sender row
	_, err = stub.ReplaceRow("InsuranceAmount", shim.Row{
		Columns : []*shim.Column {
			&shim.Column{Value: &shim.Column_Bytes{Bytes:sender}},
			&shim.Column{Value: &shim.Column_Int64{Int64:BalanceAmount}},
		},
		})
	if err != nil {
		return nil, errors.New("Failed to replace sender row!")
	}
//change assets of receiver
	var columns1 []shim.Column
	col2 := shim.Column{Value: &shim.Column_Bytes{Bytes:receiver}}
	columns1 = append(columns1, col2)

	row2, err := stub.GetRow("InsuranceAmount", columns1)
	if err != nil {
		return nil, errors.New("Failed to Get sender's Balance amount")
	}

	BalAmt := row2.Columns[1].GetInt64()
	BalAmt = BalAmt + transferAmount
//replace recv row
	_, err = stub.ReplaceRow("InsuranceAmount", shim.Row{
		Columns: []*shim.Column {
			&shim.Column{Value: &shim.Column_Bytes{Bytes:receiver}},
			&shim.Column{Value: &shim.Column_Int64{Int64:BalAmt}},
		},
	})
	if err != nil {
		return nil, errors.New("Failed to replace receiver's row!")
	}

	return nil, nil
}

func (t *SimpleHealthChaincode) isCaller(stub shim.ChaincodeStubInterface, certificate []byte) (bool, error) {
	// Verify \sigma=Sign(certificate.sk, tx.Payload||tx.Binding) against certificate.vk
	fmt.Println("isCaller is Running!")

	sigma, err := stub.GetCallerMetadata()
	if err != nil {
		return false, errors.New("Failed to get Metadata")
	}
	payload, err := stub.GetPayload()
	if err != nil {
		return false, errors.New("Failed to get payload")
	}
	binding, err := stub.GetBinding()
	if err != nil {
		return false, errors.New("Failed to get binding")
	}

	fmt.Println("Certificate : [%x]", certificate)
	fmt.Println("Sigma : [%x]", sigma)
	fmt.Println("Payload : [%x]", payload)
	fmt.Println("Binding : [%x]", binding)

	ok, err := stub.VerifySignature(
		certificate,
		sigma,
		append(payload, binding...),
	)
	if err != nil {
		return ok, errors.New("Failed Verifying signatures")
	}
	if !ok {
		fmt.Println("Signatures Does Not Match!")
	}
	fmt.Println("finished isCaller")
	
	return ok, err
}


// Invoke is our entry point to invoke a chaincode function
func (t *SimpleHealthChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "approve" {													//initialize the chaincode state, used as reset
		return t.approve(stub, args)
	} else if function == "transfer"{
		return t.transfer(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleHealthChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {											//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query: " + function)
}

func (t *SimpleHealthChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){
	
	if len(args) != 1 {
		return nil, errors.New("Expected 1 argument!")
	}
	applicant := []byte(args[0])
	//fmt.Println("Finding [%x]",string(applicant))

	var columns []shim.Column
	col := shim.Column{Value: &shim.Column_Bytes{Bytes: applicant}}
	columns = append(columns,col)

	row, err := stub.GetRow("InsuranceAmount",columns)
	if err != nil {
		return nil, errors.New("Cannot retrieve Rows")
	}
	
	fmt.Println("Finished Query function")
	
	rowString := fmt.Sprintf("%s", row)
	return []byte(rowString), nil
	//return row.Columns[0].GetBytes(), nil

}
