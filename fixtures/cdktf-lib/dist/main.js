"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.MyResource = exports.MyStack = void 0;
const provider_null_1 = require("@cdktf/provider-null");
const cdktf_1 = require("cdktf");
class MyStack extends cdktf_1.TerraformStack {
    constructor(scope, id, backend) {
        super(scope, id);
        new cdktf_1.HttpBackend(this, backend);
        new provider_null_1.provider.NullProvider(this, "null", {});
    }
}
exports.MyStack = MyStack;
class MyResource extends cdktf_1.TerraformElement {
    constructor(scope, id) {
        super(scope, id);
        new provider_null_1.dataNullDataSource.DataNullDataSource(this, "Resource", {
            inputs: {
                example: "example",
            },
        });
    }
}
exports.MyResource = MyResource;
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoibWFpbi5qcyIsInNvdXJjZVJvb3QiOiIiLCJzb3VyY2VzIjpbIi4uL21haW4udHMiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6Ijs7O0FBQUEsd0RBQW9FO0FBRXBFLGlDQUtlO0FBRWYsTUFBYSxPQUFRLFNBQVEsc0JBQWM7SUFDekMsWUFBWSxLQUFnQixFQUFFLEVBQVUsRUFBRSxPQUEwQjtRQUNsRSxLQUFLLENBQUMsS0FBSyxFQUFFLEVBQUUsQ0FBQyxDQUFDO1FBQ2pCLElBQUksbUJBQVcsQ0FBQyxJQUFJLEVBQUUsT0FBTyxDQUFDLENBQUM7UUFDL0IsSUFBSSx3QkFBUSxDQUFDLFlBQVksQ0FBQyxJQUFJLEVBQUUsTUFBTSxFQUFFLEVBQUUsQ0FBQyxDQUFDO0lBQzlDLENBQUM7Q0FDRjtBQU5ELDBCQU1DO0FBRUQsTUFBYSxVQUFXLFNBQVEsd0JBQWdCO0lBQzlDLFlBQVksS0FBZ0IsRUFBRSxFQUFVO1FBQ3RDLEtBQUssQ0FBQyxLQUFLLEVBQUUsRUFBRSxDQUFDLENBQUM7UUFDakIsSUFBSSxrQ0FBa0IsQ0FBQyxrQkFBa0IsQ0FBQyxJQUFJLEVBQUUsVUFBVSxFQUFFO1lBQzFELE1BQU0sRUFBRTtnQkFDTixPQUFPLEVBQUUsU0FBUzthQUNuQjtTQUNGLENBQUMsQ0FBQztJQUNMLENBQUM7Q0FDRjtBQVRELGdDQVNDIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0IHsgcHJvdmlkZXIsIGRhdGFOdWxsRGF0YVNvdXJjZSB9IGZyb20gXCJAY2RrdGYvcHJvdmlkZXItbnVsbFwiO1xuaW1wb3J0IHsgQ29uc3RydWN0IH0gZnJvbSBcImNvbnN0cnVjdHNcIjtcbmltcG9ydCB7XG4gIFRlcnJhZm9ybVN0YWNrLFxuICBUZXJyYWZvcm1FbGVtZW50LFxuICBIdHRwQmFja2VuZCxcbiAgSHR0cEJhY2tlbmRDb25maWcsXG59IGZyb20gXCJjZGt0ZlwiO1xuXG5leHBvcnQgY2xhc3MgTXlTdGFjayBleHRlbmRzIFRlcnJhZm9ybVN0YWNrIHtcbiAgY29uc3RydWN0b3Ioc2NvcGU6IENvbnN0cnVjdCwgaWQ6IHN0cmluZywgYmFja2VuZDogSHR0cEJhY2tlbmRDb25maWcpIHtcbiAgICBzdXBlcihzY29wZSwgaWQpO1xuICAgIG5ldyBIdHRwQmFja2VuZCh0aGlzLCBiYWNrZW5kKTtcbiAgICBuZXcgcHJvdmlkZXIuTnVsbFByb3ZpZGVyKHRoaXMsIFwibnVsbFwiLCB7fSk7XG4gIH1cbn1cblxuZXhwb3J0IGNsYXNzIE15UmVzb3VyY2UgZXh0ZW5kcyBUZXJyYWZvcm1FbGVtZW50IHtcbiAgY29uc3RydWN0b3Ioc2NvcGU6IENvbnN0cnVjdCwgaWQ6IHN0cmluZykge1xuICAgIHN1cGVyKHNjb3BlLCBpZCk7XG4gICAgbmV3IGRhdGFOdWxsRGF0YVNvdXJjZS5EYXRhTnVsbERhdGFTb3VyY2UodGhpcywgXCJSZXNvdXJjZVwiLCB7XG4gICAgICBpbnB1dHM6IHtcbiAgICAgICAgZXhhbXBsZTogXCJleGFtcGxlXCIsXG4gICAgICB9LFxuICAgIH0pO1xuICB9XG59XG4iXX0=