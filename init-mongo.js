// db.createUser(
//     {
//       user: "root",
//       pwd: "83000000",
//       roles: [
//         {
//           role: "readWrite",
//           db: "admin"
//         }
//       ]
//     }
//   );
var admin = db.getSiblingDB("admin");

// Create a new user with the desired username and password
admin.createUser({
  user: "root",
  pwd: "83000000",
  roles: [{ role: "readWrite", db: "admin" }]
});