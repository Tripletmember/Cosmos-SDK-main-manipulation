# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [internal/store/types/types.proto](#internal/store/types/types.proto)
    - [InternalObject](#cosmosSdkCrud.internal.store.types.v1beta1.InternalObject)
    - [InternalSecondaryKey](#cosmosSdkCrud.internal.store.types.v1beta1.InternalSecondaryKey)
    - [indexList](#cosmosSdkCrud.internal.store.types.v1beta1.indexList)
  
- [internal/store/types/types_test.proto](#internal/store/types/types_test.proto)
    - [TestObject](#cosmosSdkCrud.internal.store.types.v1beta1.TestObject)
  
- [Scalar Value Types](#scalar-value-types)



<a name="internal/store/types/types.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/store/types/types.proto



<a name="cosmosSdkCrud.internal.store.types.v1beta1.InternalObject"></a>

### InternalObject
InternalObject defines a structure that can be saved in the crud store


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| primary_key | [bytes](#bytes) |  | PrimaryKey is the unique id that identifies the object |
| secondary_keys | [InternalSecondaryKey](#cosmosSdkCrud.internal.store.types.v1beta1.InternalSecondaryKey) | repeated | SecondaryKeys is an array containing the secondary keys used to map the object |






<a name="cosmosSdkCrud.internal.store.types.v1beta1.InternalSecondaryKey"></a>

### InternalSecondaryKey



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int32](#int32) |  | TODO: FIXME what protobuf type to use for byte? |
| value | [bytes](#bytes) |  |  |






<a name="cosmosSdkCrud.internal.store.types.v1beta1.indexList"></a>

### indexList
indexList


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| indexes | [bytes](#bytes) | repeated | Indexes |





 

 

 

 



<a name="internal/store/types/types_test.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/store/types/types_test.proto



<a name="cosmosSdkCrud.internal.store.types.v1beta1.TestObject"></a>

### TestObject
TestObject is a mock object used to test the store


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| TestPrimaryKey | [bytes](#bytes) |  | TestPrimaryKey is a primary key |
| TestSecondaryKeyA | [bytes](#bytes) |  | TestSecondaryKeyA is secondary key number one |
| TestSecondaryKeyB | [bytes](#bytes) |  | TestSecondaryKeyB is secondary key number two |





 

 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

