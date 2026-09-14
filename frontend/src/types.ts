export type Product = { id:number; name:string; slug:string; category:string; price:number; rating:number; reviews:number; stock:number; color:string; accent:string; description:string; featured:boolean }
export type CartLine = Product & { quantity:number }
