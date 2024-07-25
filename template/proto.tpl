syntax = "proto3";
package api.{{.PackageName}}.v1;

option go_package = "pb/{{.PackageName}}Pb;{{.PackageName}}Pb";
// option csharp_namespace = "Assets.Script.Reborn.Proto.Cbpb";


service {{ .ServiceName }} {
  // create {{ .TableName }}
  rpc Create(Create{{ .TableName }}Req) returns (Create{{ .TableName }}Rsp) {}

  // delete {{ .TableName }} by id
  rpc DeleteByID(Delete{{ .TableName }}ByIDReq) returns (Delete{{ .TableName }}ByIDRsp) {}

  // delete {{ .TableName }} by batch id
  rpc DeleteByIDs(Delete{{ .TableName }}ByIDsReq) returns (Delete{{ .TableName }}ByIDsRsp) {}

  // update {{ .TableName }} by id
  rpc UpdateByID(Update{{ .TableName }}ByIDReq) returns (Update{{ .TableName }}ByIDRsp) {}

  // get {{ .TableName }} by id
  rpc GetByID(Get{{ .TableName }}ByIDReq) returns (Get{{ .TableName }}ByIDRsp) {}

  // list of {{ .TableName }} by batch id
  rpc ListByIDs(List{{ .TableName }}ByIDsReq) returns (List{{ .TableName }}ByIDsRsp) {}
}

// create request
message Create{{ .TableName }}Req {
      {{- range .ProtoColumns}}
      {{ .ColumnType }} {{ .FieldName }} = {{ .ColumnNum }};
      {{- end}}
}

// create response
message Create{{ .TableName }}Rsp {
}

// delete request
message Delete{{ .TableName }}ByIDReq {
  {{.PrimaryKeyProtoType}} {{.PrimaryKey}} = 1;
}

// delete response
message Delete{{ .TableName }}ByIDRsp {

}

// delete list request
message Delete{{ .TableName }}ByIDsReq {
  repeated {{.PrimaryKeyProtoType}} {{.PrimaryKey}}s = 1;
}

// delete list response
message Delete{{ .TableName }}ByIDsRsp {

}

// update request
message Update{{ .TableName }}ByIDReq {
      {{- range .ProtoColumns}}
      {{ .ColumnType }} {{ .FieldName }} = {{ .ColumnNum }};
      {{- end}}
}

// update response
message Update{{ .TableName }}ByIDRsp {

}

// get one request
message Get{{ .TableName }}ByIDReq {
  {{.PrimaryKeyProtoType}} {{.PrimaryKey}} = 1;
}

// get one response
message Get{{ .TableName }}ByIDRsp {
      {{- range .ProtoColumns}}
      {{ .ColumnType }} {{ .FieldName }} = {{ .ColumnNum }};
      {{- end}}
}


// get list request
message List{{ .TableName }}ByIDsReq {
    repeated int64 {{.PrimaryKey}}s = 1;
}

// get list response
message List{{ .TableName }}ByIDsRsp {
    repeated {{ .TableName }} list = 1;
}


message {{ .TableName }} {
    {{- range .ProtoColumns}}
    {{ .ColumnType }} {{ .FieldName }} = {{ .ColumnNum }};
    {{- end}}
}
